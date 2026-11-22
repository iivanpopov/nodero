use crate::context::{Context, Slave};
use proto::codec::decode_payload;
use proto::proto::envelope::Payload;
use shared::transport::{read_frame, write_frame};
use std::{net::SocketAddr, sync::Arc};
use tokio::{
    net::{TcpStream, tcp::OwnedReadHalf},
    sync::{RwLock, mpsc},
    time::Instant,
};
use tracing::{Instrument, debug, error, info, instrument, warn};

#[instrument(skip_all, fields(%source_addr))]
pub async fn handle_udp_payload(payload: Payload, source_addr: SocketAddr, ctx: &Arc<Context>) {
    match payload {
        Payload::DiscoverResponse(response) => {
            if response.device_id.is_empty() {
                return;
            }

            debug!(device_id = %response.device_id, "heartbeat received");

            let mut slaves = ctx.slaves.write().await;

            if let Some(slave) = slaves.get(&response.device_id).cloned() {
                drop(slaves);
                slave.write().await.last_heartbeat = Instant::now();
                return;
            }

            info!(device_id = %response.device_id, "discover response");

            let device_id = response.device_id;
            let client = spawn_device_connection(source_addr, device_id.clone(), Arc::clone(ctx));
            slaves.insert(device_id, client);
        }
        Payload::DiscoverRequest(_) => warn!("unexpected DiscoverRequest"),
    }
}

#[instrument(name = "tcp_connection", skip_all, fields(%addr, %device_id))]
fn spawn_device_connection(
    addr: SocketAddr,
    device_id: String,
    ctx: Arc<Context>,
) -> Arc<RwLock<Slave>> {
    let (tx, mut rx) = mpsc::channel(8);
    let slave = Arc::new(RwLock::new(Slave {
        addr,
        tx,
        last_heartbeat: Instant::now(),
    }));
    let session = Arc::downgrade(&slave);

    tokio::spawn(
        async move {
            match TcpStream::connect(addr).await {
                Ok(stream) => {
                    let (reader, mut writer) = stream.into_split();

                    let write = async {
                        while let Some(data) = rx.recv().await {
                            if let Err(error) = write_frame(&mut writer, &data).await {
                                error!(%error, "failed to write frame");
                                break;
                            }
                        }
                    };

                    tokio::select! {
                        () = read_tcp_frames(reader, &device_id, &ctx) => {},
                        () = write => {},
                    }
                }
                Err(error) => error!(%error, "failed to connect via tcp"),
            }

            let mut slaves = ctx.slaves.write().await;
            if slaves
                .get(&device_id)
                .is_some_and(|current| Arc::as_ptr(current) == session.as_ptr())
            {
                slaves.remove(&device_id);
            }
        }
        .in_current_span(),
    );

    slave
}

async fn read_tcp_frames(mut reader: OwnedReadHalf, device_id: &str, ctx: &Arc<Context>) {
    loop {
        let frame = match read_frame(&mut reader).await {
            Ok(frame) => frame,
            Err(error) => {
                warn!(%error, "failed reading frame");
                return;
            }
        };

        match decode_payload(&frame) {
            Ok(payload) => handle_tcp_payload(payload, device_id, ctx).await,
            Err(error) => {
                warn!(%error, "invalid tcp message");
                debug!(bytes = ?frame.get(..16).unwrap_or(&frame), "invalid message bytes");
            }
        }
    }
}

async fn handle_tcp_payload(payload: Payload, _device_id: &str, _ctx: &Arc<Context>) {
    match payload {
        _ => {
            warn!("unexpected payload");
        }
    }
}
