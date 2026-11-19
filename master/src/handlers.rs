use crate::context::{Context, TcpClient};
use proto::codec::decode_payload;
use proto::proto::envelope::Payload;
use shared::transport::read_frame;
use std::{net::SocketAddr, sync::Arc};
use tokio::{net::TcpStream, sync::mpsc};
use tracing::{debug, error, info, instrument, warn};

#[instrument(skip_all, fields(%source_addr))]
pub async fn handle_udp_payload(payload: Payload, source_addr: SocketAddr, ctx: &Arc<Context>) {
    match payload {
        Payload::DiscoverRequest(request) => {
            info!(device_id = %request.device_id, "discover request");

            let ctx = Arc::clone(ctx);

            tokio::spawn(async move {
                connect_to_device(source_addr, request.device_id, &ctx).await;
            });
        }
    }
}

#[instrument(skip_all, fields(%addr, %device_id))]
async fn connect_to_device(addr: SocketAddr, device_id: String, ctx: &Arc<Context>) {
    match TcpStream::connect(addr).await {
        Ok(stream) => handle_tcp_connection(stream, addr, device_id, ctx).await,
        Err(error) => error!(%error, "failed to connect via tcp"),
    }
}

async fn handle_tcp_connection(
    stream: TcpStream,
    addr: SocketAddr,
    device_id: String,
    ctx: &Arc<Context>,
) {
    let (mut reader, _writer) = stream.into_split();
    let (tx, _rx) = mpsc::channel::<Vec<u8>>(64);

    ctx.tcp_clients.write().await.insert(
        device_id.clone(),
        TcpClient {
            addr: addr,
            tx: tx.clone(),
        },
    );

    loop {
        let frame = match read_frame(&mut reader).await {
            Ok(frame) => frame,
            Err(error) => {
                warn!(%error, "failed reading frame");
                break;
            }
        };

        match decode_payload(&frame) {
            Ok(payload) => handle_tcp_payload(payload, &device_id, &ctx).await,
            Err(error) => {
                warn!(%error, "invalid tcp message");
                debug!(bytes = ?frame.get(..16).unwrap_or(&frame), "invalid message bytes");
                continue;
            }
        };
    }

    let mut device_senders = ctx.tcp_clients.write().await;
    if device_senders
        .get(&device_id)
        .is_some_and(|sender| sender.tx.same_channel(&tx))
    {
        device_senders.remove(&device_id);
    }
}

async fn handle_tcp_payload(payload: Payload, _device_id: &str, _ctx: &Arc<Context>) {
    match payload {
        Payload::DiscoverRequest(_) => {
            warn!("unexpected DiscoverRequest over tcp");
        }
    }
}
