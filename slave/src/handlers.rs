use std::{net::SocketAddr, sync::Arc};

use prost::Message;
use proto::{
    codec::decode_payload,
    proto::{DiscoverRequest, Envelope, envelope::Payload},
};
use shared::transport::{read_frame, write_frame};
use tokio::{net::TcpStream, sync::mpsc};
use tracing::{debug, instrument, warn};

use crate::context::Context;

#[instrument(skip_all, fields(%addr))]
pub async fn handle_tcp_connection(stream: TcpStream, addr: SocketAddr, ctx: &Arc<Context>) {
    let (mut reader, mut writer) = stream.into_split();
    let (tx, _rx) = mpsc::channel::<Vec<u8>>(64);

    *ctx.tcp_sender.lock().await = Some(tx.clone());

    let data = Envelope::from(DiscoverRequest {
        device_id: ctx.device_id.clone(),
    })
    .encode_to_vec();
    write_frame(&mut writer, &data).await.unwrap();

    loop {
        let frame = match read_frame(&mut reader).await {
            Ok(frame) => frame,
            Err(error) => {
                warn!(%error, "failed reading frame");
                break;
            }
        };

        match decode_payload(&frame) {
            Ok(payload) => handle_tcp_payload(payload, &ctx).await,
            Err(error) => {
                warn!(%error, "invalid tcp message");
                debug!(bytes = ?frame.get(..16).unwrap_or(&frame), "invalid message bytes");
                continue;
            }
        };
    }

    let mut sender = ctx.tcp_sender.lock().await;

    if sender
        .as_ref()
        .is_some_and(|current| current.same_channel(&tx))
    {
        *sender = None;
    }
}

#[instrument(skip_all)]
pub async fn handle_tcp_payload(payload: Payload, _ctx: &Arc<Context>) {
    match payload {
        Payload::DiscoverRequest(_) => {
            warn!("unexpected DiscoverRequest over tcp");
        }
    }
}
