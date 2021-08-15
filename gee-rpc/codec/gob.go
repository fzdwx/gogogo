package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec Codec 的 gob实现
type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer // 缓冲区
	dec  *gob.Decoder  // 编解码器
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

// NewGobCodecFunc GobCodec构造函数
func NewGobCodecFunc(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// impl Codec

func (c GobCodec) Close() error {
	return c.conn.Close()
}

func (c GobCodec) ReadHeader(header *RequestHeader) error {
	return c.dec.Decode(header)
}

func (c GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c GobCodec) Write(header *RequestHeader, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err := c.enc.Encode(header); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil

}
