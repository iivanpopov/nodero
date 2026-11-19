use std::{env, error::Error, path::PathBuf};

fn main() -> Result<(), Box<dyn Error>> {
    let proto_root = PathBuf::from(env::var("CARGO_MANIFEST_DIR")?).join("../../proto");

    println!("cargo:rerun-if-changed={}", proto_root.display());
    println!("cargo:rerun-if-env-changed=PROTOC");
    println!("cargo:rerun-if-env-changed=PROTOC_INCLUDE");

    prost_build::compile_protos(
        &[
            proto_root.join("discovery.proto"),
            proto_root.join("envelope.proto"),
        ],
        &[proto_root],
    )?;

    Ok(())
}
