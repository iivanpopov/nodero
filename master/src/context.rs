use std::{collections::HashMap, net::SocketAddr, sync::Arc};
use tokio::{
    net::UdpSocket,
    sync::{RwLock, mpsc::Sender},
    time::Instant,
};

pub struct Slave {
    pub addr: SocketAddr,
    pub tx: Sender<Vec<u8>>,
    pub last_heartbeat: Instant,
}

pub struct Context {
    pub udp: UdpSocket,
    pub slaves: RwLock<HashMap<String, Arc<RwLock<Slave>>>>,
}
