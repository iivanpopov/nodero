use tokio::{
    net::{TcpListener, UdpSocket},
    sync::{Mutex, mpsc::Sender},
};

pub struct Context {
    pub device_id: String,
    pub udp: UdpSocket,
    pub tcp: TcpListener,
    pub tcp_sender: Mutex<Option<Sender<Vec<u8>>>>,
}
