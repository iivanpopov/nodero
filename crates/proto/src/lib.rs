pub mod codec;

pub mod proto {
    include!(concat!(env!("OUT_DIR"), "/nodero.rs"));

    impl From<DiscoverRequest> for Envelope {
        fn from(from: DiscoverRequest) -> Self {
            Envelope {
                payload: Some(envelope::Payload::DiscoverRequest(from)),
            }
        }
    }

    impl From<DiscoverResponse> for Envelope {
        fn from(from: DiscoverResponse) -> Self {
            Envelope {
                payload: Some(envelope::Payload::DiscoverResponse(from)),
            }
        }
    }
}
