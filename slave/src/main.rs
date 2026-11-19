use prost::Message;
use proto::proto::{DiscoverRequest, Envelope};
use shared::{
    config::{get_device_id, get_master_port, get_slave_port},
    logger,
};
use slave::{context::Context, handlers::handle_tcp_connection};
use std::{error::Error, sync::Arc};
use tokio::{
    net::{TcpListener, UdpSocket},
    sync::Mutex,
};
use tracing::{error, info};

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error + Send + Sync>> {
    let log_dir = logger::get_log_dir()?;
    let _guard = logger::init(log_dir)?;

    let master_port = get_master_port();

    let slave_port = get_slave_port();
    let slave_addr = format!("0.0.0.0:{slave_port}");

    let ctx = Arc::new(Context {
        device_id: get_device_id()?,
        udp: UdpSocket::bind(&slave_addr).await?,
        tcp: TcpListener::bind(&slave_addr).await?,
        tcp_sender: Mutex::new(None),
    });

    ctx.udp.set_broadcast(true)?;

    let broadcast_addr = format!("255.255.255.255:{master_port}");
    let message = Envelope::from(DiscoverRequest {
        device_id: ctx.device_id.clone(),
    })
    .encode_to_vec();
    ctx.udp.send_to(&message, broadcast_addr).await?;

    let tcp_task = tokio::spawn(async move {
        let ctx = Arc::clone(&ctx);

        let Ok(local_addr) = ctx.tcp.local_addr() else {
            error!("failed to get tcp local address");
            return;
        };

        info!("listening tcp {local_addr}");

        loop {
            match ctx.tcp.accept().await {
                Ok((stream, addr)) => handle_tcp_connection(stream, addr, &ctx).await,
                Err(error) => error!(%error, "failed to accept tcp connection"),
            }
        }
    });

    tcp_task.await?;

    Ok(())
}
