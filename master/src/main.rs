use master::{context::Context, handlers::handle_udp_payload};
use prost::Message;
use proto::{
    codec::decode_payload,
    proto::{DiscoverRequest, Envelope},
};
use shared::{
    config::{get_master_port, get_slave_port},
    logger,
};
use std::{collections::HashMap, error::Error, sync::Arc, time::Duration};
use tokio::{
    net::UdpSocket,
    sync::RwLock,
    time::{MissedTickBehavior, interval},
};
use tracing::{debug, error, info, warn};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error + Send + Sync>> {
    let log_dir = logger::get_log_dir()?;
    let _guard = logger::init(log_dir)?;

    let master_port = get_master_port();
    let slave_port = get_slave_port();
    let master_addr = format!("0.0.0.0:{master_port}");
    let broadcast_addr = format!("255.255.255.255:{slave_port}");

    let ctx = Arc::new(Context {
        udp: UdpSocket::bind(&master_addr).await?,
        slaves: RwLock::new(HashMap::new()),
    });
    ctx.udp.set_broadcast(true)?;

    let broadcast_task = tokio::spawn({
        let ctx = Arc::clone(&ctx);

        async move {
            let message = Envelope::from(DiscoverRequest {}).encode_to_vec();
            let mut timer = interval(Duration::from_secs(3));
            timer.set_missed_tick_behavior(MissedTickBehavior::Skip);

            loop {
                timer.tick().await;
                if let Err(error) = ctx.udp.send_to(&message, &broadcast_addr).await {
                    error!(%error, "failed to broadcast discovery request");
                }

                let clients = ctx.slaves.read().await;

                for (device_id, client) in clients.iter() {
                    let client = client.read().await;

                    if client.last_heartbeat.elapsed() >= Duration::from_secs(5) {
                        warn!(%device_id, addr = %client.addr, "no heartbeat");
                    }
                }
            }
        }
    });

    let udp_task = tokio::spawn({
        let ctx = Arc::clone(&ctx);

        async move {
            let Ok(local_addr) = ctx.udp.local_addr() else {
                error!("failed to get udp local address");
                return;
            };

            info!("listening udp {local_addr}");

            loop {
                let mut buf = [0u8; 1024];

                match ctx.udp.recv_from(&mut buf).await {
                    Ok((len, addr)) => {
                        let bytes = &buf[..len];

                        match decode_payload(bytes) {
                            Ok(payload) => handle_udp_payload(payload, addr, &ctx).await,
                            Err(error) => {
                                warn!(%addr, %error, "invalid message");
                                debug!(%addr, bytes = ?bytes.get(..16).unwrap_or(bytes), "invalid message bytes");
                            }
                        }
                    }
                    Err(error) => error!(%error, "failed receiving datagram"),
                }
            }
        }
    });

    tokio::select! {
        result = udp_task => result?,
        result = broadcast_task => result?,
    }

    Ok(())
}
