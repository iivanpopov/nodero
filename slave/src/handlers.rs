use std::net::SocketAddr;

use prost::Message;
use proto::{
    codec::decode_payload,
    proto::{DiscoverResponse, Envelope, envelope::Payload},
};
use shared::transport::{read_frame, write_frame};
use tokio::{
    net::{TcpStream, tcp::OwnedReadHalf},
    sync::mpsc,
};
use tracing::{debug, error, instrument, warn};

use crate::context::Context;

#[instrument(skip_all, fields(%source_addr))]
pub async fn handle_udp_payload(payload: Payload, source_addr: SocketAddr, ctx: &Context) {
    match payload {
        Payload::DiscoverRequest(_) => {
            let message = Envelope::from(DiscoverResponse {
                device_id: ctx.device_id.clone(),
            })
            .encode_to_vec();

            if let Err(error) = ctx.udp.send_to(&message, source_addr).await {
                error!(%error, "failed to send discovery response");
            }
        }
        _ => warn!("unexpected payload"),
    }
}

#[instrument(name = "tcp_connection", skip_all, fields(%addr))]
pub async fn handle_tcp_connection(stream: TcpStream, addr: SocketAddr, ctx: &Context) {
    let (reader, mut writer) = stream.into_split();
    let (tx, mut rx) = mpsc::channel(8);

    *ctx.tx.lock().await = Some(tx);

    let write = async {
        while let Some(data) = rx.recv().await {
            if let Err(error) = write_frame(&mut writer, &data).await {
                error!(%error, "failed to write frame");
                break;
            }
        }
    };

    tokio::select! {
        () = read_tcp(reader, ctx) => {},
        () = write => {},
    }

    *ctx.tx.lock().await = None;
}

async fn read_tcp(mut reader: OwnedReadHalf, ctx: &Context) {
    loop {
        let frame = match read_frame(&mut reader).await {
            Ok(frame) => frame,
            Err(error) => {
                warn!(%error, "failed reading frame");
                break;
            }
        };

        match decode_payload(&frame) {
            Ok(payload) => handle_tcp_payload(payload, ctx).await,
            Err(error) => {
                warn!(%error, "invalid tcp message");
                debug!(bytes = ?frame.get(..16).unwrap_or(&frame), "invalid message bytes");
            }
        }
    }
}

pub async fn handle_tcp_payload(payload: Payload, _ctx: &Context) {
    match payload {
        _ => {
            warn!("unexpected payload");
        }
    }
}
