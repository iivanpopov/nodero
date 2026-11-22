use proto::codec::decode_payload;
use shared::{
    config::{get_device_id, get_slave_port},
    logger,
};
use slave::{
    context::Context,
    handlers::{handle_tcp_connection, handle_udp_payload},
};
use std::{error::Error, sync::Arc};
use tokio::{
    net::{TcpListener, UdpSocket},
    sync::Mutex,
};
use tracing::{debug, error, info, warn};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error + Send + Sync>> {
    let log_dir = logger::get_log_dir()?;
    let _guard = logger::init(log_dir)?;

    let slave_port = get_slave_port();
    let slave_addr = format!("0.0.0.0:{slave_port}");

    let tcp = TcpListener::bind(&slave_addr).await?;

    let context = Arc::new(Context {
        device_id: get_device_id()?,
        udp: UdpSocket::bind(&slave_addr).await?,
        tx: Mutex::new(None),
    });

    let ctx = Arc::clone(&context);
    let udp_task = tokio::spawn(async move {
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
    });

    let ctx = Arc::clone(&context);
    let tcp_task = tokio::spawn(async move {
        let Ok(local_addr) = tcp.local_addr() else {
            error!("failed to get tcp local address");
            return;
        };

        info!("listening tcp {local_addr}");

        loop {
            match tcp.accept().await {
                Ok((stream, addr)) => handle_tcp_connection(stream, addr, &ctx).await,
                Err(error) => error!(%error, "failed to accept tcp connection"),
            }
        }
    });

    tokio::select! {
        result = tcp_task => result?,
        result = udp_task => result?,
    }

    Ok(())
}
