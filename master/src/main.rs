use master::{context::Context, handlers::handle_udp_payload};
use proto::codec::decode_payload;
use shared::{config::get_master_port, logger};
use std::{collections::HashMap, error::Error, sync::Arc};
use tokio::{net::UdpSocket, sync::RwLock};
use tracing::{debug, error, info, warn};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error + Send + Sync>> {
    let log_dir = logger::get_log_dir()?;
    let _guard = logger::init(log_dir)?;

    let master_port = get_master_port();
    let master_addr = format!("0.0.0.0:{master_port}");

    let ctx = Arc::new(Context {
        udp: UdpSocket::bind(&master_addr).await?,
        tcp_clients: RwLock::new(HashMap::new()),
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

    let _ = udp_task.await;

    Ok(())
}
