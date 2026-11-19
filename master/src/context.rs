use std::collections::HashMap;

use tokio::{
    net::UdpSocket,
    sync::{RwLock, mpsc::Sender},
};

pub struct Context {
    pub udp: UdpSocket,
    pub tcp_senders: RwLock<HashMap<String, Sender<Vec<u8>>>>,
}
