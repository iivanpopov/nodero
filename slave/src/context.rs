use tokio::{
    net::UdpSocket,
    sync::{Mutex, mpsc::Sender},
};

pub struct Context {
    pub device_id: String,
    pub udp: UdpSocket,
    pub tx: Mutex<Option<Sender<Vec<u8>>>>,
}
