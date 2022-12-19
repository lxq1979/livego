package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"time"

	"github.com/gwuhaolin/livego/utils/pio"
)

var (
	timeout = 5 * time.Second
)

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

func hsMakeDigest(key []byte, src []byte, gap int) (dst []byte) {
	h := hmac.New(sha256.New, key)
	if gap <= 0 {
		h.Write(src)
	} else {
		h.Write(src[:gap])
		h.Write(src[gap+32:])
	}
	return h.Sum(nil)
}

func hsCalcDigestPos(p []byte, base int) (pos int) {
	for i := 0; i < 4; i++ {
		pos += int(p[base+i])
	}
	pos = (pos % 728) + base + 4
	return
}

func hsFindDigest(p []byte, key []byte, base int) int {
	gap := hsCalcDigestPos(p, base)
	digest := hsMakeDigest(key, p, gap)
	if bytes.Compare(p[gap:gap+32], digest) != 0 {
		return -1
	}
	return gap
}

func hsParse1(p []byte, peerkey []byte, key []byte) (ok bool, digest []byte) {
	var pos int
	if pos = hsFindDigest(p, peerkey, 772); pos == -1 {
		if pos = hsFindDigest(p, peerkey, 8); pos == -1 {
			return
		}
	}
	ok = true
	digest = hsMakeDigest(key, p[pos:pos+32], -1)
	return
}

func hsCreate01(p []byte, time uint32, ver uint32, key []byte) {
	p[0] = 3
	p1 := p[1:]
	rand.Read(p1[8:])
	pio.PutU32BE(p1[0:4], time)
	pio.PutU32BE(p1[4:8], ver)
	gap := hsCalcDigestPos(p1, 8)
	digest := hsMakeDigest(key, p1, gap)
	copy(p1[gap:], digest)
}

func hsCreate2(p []byte, key []byte) {
	rand.Read(p)
	gap := len(p) - 32
	digest := hsMakeDigest(key, p, gap)
	copy(p[gap:], digest)
}

func (conn *Conn) HandshakeClient() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]

	C0[0] = 3
	// > C0C1
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.rw.Write(C0C1); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if err = conn.rw.Flush(); err != nil {
		return
	}

	// < S0S1S2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = io.ReadFull(conn.rw, S0S1S2); err != nil {
		return
	}

	S1 := S0S1S2[1 : 1536+1]
	if ver := pio.U32BE(S1[4:8]); ver != 0 {
		C2 = S1
	} else {
		C2 = S1
	}

	// > C2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = conn.rw.Write(C2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Time{})
	return
}

/**
LXQ: RTMP握手的規則，參【RTMP 协议：为什么直播推流协议都爱用它】,之握手篇
C0,S0 1byte version 3
C1,S1  1536byte 包括
	time 4byte
	zero 4byte must 全 0
	random 1528bytes 本字段可以包含任何值。由于握手的双方需要区分另一端，此字段填充的数据必须足够随机（以防止与其他握手端混淆）。不过没有必要为此使用加密数据或动态数据。

C2 和 S2（1536 bytes）
	time（4 bytes）：本字段表示对端发送的时间戳（对C2来说是S1 ,对S2来说是C1）
	time2（4 bytes）：本字段表示接收对端发送过来的握手包的时间戳
	random（1528 bytes）：本字段包含对端发送过来的随机数据（对C2来说是S1，对S2来说是C1）

整個過程
为了方便开发，在实现上，我们选用以下握手流程，这样服务端可以连续发送S0,S1和S2：
Client--> Server : 发送一个创建流的请求(C0、C1)
Server--> Client : 返回一个流的索引号( S0、S1、S2)。
Client--> Server : 开始发送 (C2)
Client--> Server : 发送音视频数据(这些包用流的索引号来唯一标识)

*/
func (conn *Conn) HandshakeServer() (err error) {
	var random [(1 + 1536*2) * 2]byte

	C0C1C2 := random[:1536*2+1]
	C0 := C0C1C2[:1]
	C1 := C0C1C2[1 : 1536+1]
	C0C1 := C0C1C2[:1536+1]
	C2 := C0C1C2[1536+1:]

	S0S1S2 := random[1536*2+1:]
	S0 := S0S1S2[:1]
	S1 := S0S1S2[1 : 1536+1]
	S0S1 := S0S1S2[:1536+1]
	S2 := S0S1S2[1536+1:]

	// < C0C1
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	//LXQ:接收client發過來的 C0C1
	if _, err = io.ReadFull(conn.rw, C0C1); err != nil {
		return
	}

	conn.Conn.SetDeadline(time.Now().Add(timeout))
	//lxq: 在 C0 中，该字段标识了客户端请求的 RTMP 版本。
	//在 S0 中，这个字段是服务器的 RTMP 版本。这个版本被定义成 3
	if C0[0] != 3 {
		err = fmt.Errorf("rtmp: handshake version=%d invalid", C0[0])
		return
	}

	S0[0] = 3

	clitime := pio.U32BE(C1[0:4])
	srvtime := clitime
	srvver := uint32(0x0d0e0a0d)
	cliver := pio.U32BE(C1[4:8])

	if cliver != 0 {
		var ok bool
		var digest []byte
		if ok, digest = hsParse1(C1, hsClientPartialKey, hsServerFullKey); !ok {
			err = fmt.Errorf("rtmp: handshake server: C1 invalid")
			return
		}
		hsCreate01(S0S1, srvtime, srvver, hsServerPartialKey)
		hsCreate2(S2, digest)
	} else {
		copy(S1, C2)
		copy(S2, C1)
	}

	// > S0S1S2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	//LXQ:向client發S0S1S2
	if _, err = conn.rw.Write(S0S1S2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if err = conn.rw.Flush(); err != nil {
		return
	}

	//LXQ: 服務端接收 < C2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err = io.ReadFull(conn.rw, C2); err != nil {
		return
	}
	conn.Conn.SetDeadline(time.Time{})
	return
}
