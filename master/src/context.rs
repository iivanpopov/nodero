use std::{collections::HashMap, net::SocketAddr};

use tokio::{
    net::UdpSocket,
    sync::{RwLock, mpsc::Sender},
};

pub struct TcpClient {
    pub addr: SocketAddr,
    pub tx: Sender<Vec<u8>>,
}

pub struct Context {
    pub udp: UdpSocket,
    pub tcp_clients: RwLock<HashMap<String, TcpClient>>,
}
