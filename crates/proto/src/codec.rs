use crate::proto::{Envelope, envelope::Payload};
use prost::Message;
use std::{error::Error, fmt};

#[derive(Debug)]
pub enum DecodePayloadError {
    InvalidProtobuf(prost::DecodeError),
    MissingPayload,
}

impl fmt::Display for DecodePayloadError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::InvalidProtobuf(error) => write!(f, "invalid protobuf: {error}"),
            Self::MissingPayload => write!(f, "missing envelope payload"),
        }
    }
}

impl Error for DecodePayloadError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            Self::InvalidProtobuf(error) => Some(error),
            Self::MissingPayload => None,
        }
    }
}

pub fn decode_payload(bytes: &[u8]) -> Result<Payload, DecodePayloadError> {
    let envelope = Envelope::decode(bytes).map_err(DecodePayloadError::InvalidProtobuf)?;
    envelope.payload.ok_or(DecodePayloadError::MissingPayload)
}
