use std::{env, fs, io, path::PathBuf};
use uuid::{Uuid, Variant};

pub fn get_master_port() -> String {
    env::var("MASTER_PORT").unwrap_or(String::from("5005"))
}

pub fn get_slave_port() -> String {
    env::var("SLAVE_PORT").unwrap_or(String::from("5004"))
}

pub fn get_program_dir() -> PathBuf {
    env::var_os("PROGRAMDATA")
        .map(PathBuf::from)
        .unwrap_or_else(env::temp_dir)
        .join("Nodero")
}

pub fn get_device_id() -> io::Result<String> {
    let dir = get_program_dir();
    let path = dir.join("device_id.txt");

    match fs::read_to_string(&path) {
        Ok(data) => {
            if let Ok(uuid) = Uuid::parse_str(data.trim()) {
                if uuid.get_variant() == Variant::RFC4122 {
                    return Ok(uuid.to_string());
                }
            }
        }
        Err(error) if error.kind() == io::ErrorKind::NotFound => {}
        Err(error) => return Err(error),
    }

    let uuid = Uuid::new_v4().to_string();

    fs::create_dir_all(&dir)?;
    fs::write(&path, &uuid)?;

    Ok(uuid)
}
