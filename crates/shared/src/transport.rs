use tokio::{
    io::{self, AsyncRead, AsyncReadExt, AsyncWriteExt},
    net::tcp::OwnedWriteHalf,
};

pub async fn read_frame<R: AsyncRead + Unpin>(reader: &mut R) -> io::Result<Vec<u8>> {
    let len = reader.read_u32().await?;
    let mut buf = vec![0u8; len as usize];
    reader.read_exact(&mut buf).await?;
    Ok(buf)
}

pub async fn write_frame(writer: &mut OwnedWriteHalf, data: &[u8]) -> io::Result<()> {
    let len = u32::try_from(data.len()).map_err(io::Error::other)?;

    let mut buf = Vec::with_capacity(4 + data.len());

    buf.extend_from_slice(&len.to_be_bytes());
    buf.extend_from_slice(&data);

    writer.write_all(&buf).await?;

    Ok(())
}
