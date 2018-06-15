package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"math/big"
	"encoding/gob"
	"bytes"
)

var tag = "节点"

const HAND_SHAKE_MSG = "我是打洞消息"
//以下是要发送的数据，先定义数据的类型
type BlockHeader struct {
	Number *big.Int
	GasLimit *big.Int
	GasUsed *big.Int
	Time *big.Int
}
type Transaction struct{
	from string
	to string
	value int
}
type Block struct {
	Header *BlockHeader
	Transactions []*Transaction
	Sig string
}
// 将区块序列化成字节数组.这是是方便我们进行数据的传输
func (block *Block) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}
// 反序列化，这里是方便我们收到数据进行反序列化
func DeserializeBlock(blockBytes []byte) *Block {

	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
//定义我们要发送的数据
var blockHeader = &BlockHeader{big.NewInt(1),big.NewInt(1000),big.NewInt(1000),big.NewInt(time.Now().Unix())}
var transaction = &Transaction{"xiaohong","laowang",10}
var transactions []*Transaction



func main() {
	transactions  = append(transactions,transaction)
	var block = Block{blockHeader,transactions,"sign_string"}

	// 当前进程标记字符串,便于显示
	//tag = os.Args[1]
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 9982} // 注意端口必须固定
	dstAddr := &net.UDPAddr{IP: net.ParseIP("192.168.2.240"), Port: 9981}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	if _, err = conn.Write([]byte("hello, I'm new peer:" + tag)); err != nil {
		log.Panic(err)
	}
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("error during read: %s", err)
	}
	conn.Close()
	anotherPeer := parseAddr(string(data[:n]))
	fmt.Printf("local:%s server:%s another:%s\n", srcAddr, remoteAddr, anotherPeer.String())

	// 开始打洞
	bidirectionHole(srcAddr, &anotherPeer,block)

}

func parseAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}
//nat打洞的程序
func bidirectionHole(srcAddr *net.UDPAddr, anotherAddr *net.UDPAddr,block Block) {
	conn, err := net.DialUDP("udp", srcAddr, anotherAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	// 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
	if _, err = conn.Write([]byte(HAND_SHAKE_MSG)); err != nil {
		log.Println("send handshake:", err)
	}
	//每隔5秒钟发送一次数据
	go func() {
		for {
			time.Sleep(5 * time.Second)
			//if _, err = conn.Write([]byte("from [" + tag + "]")); err != nil {
			//	log.Println("send msg fail", err)
			//}
			if _, err = conn.Write(block.Serialize()); err != nil {
				log.Println("send msg fail", err)
			}
		}
	}()
	//循坏来不断接受数据
	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		Dblock := DeserializeBlock(data[:n])
		if err != nil {
			log.Printf("error during read: %s\n", err)
		} else {
			log.Printf("收到数据:%s\n", data[:n])
		}
		fmt.Println("---------------收到数据------------")
		fmt.Println(Dblock.Transactions)
		fmt.Println(Dblock.Header)
		fmt.Println(Dblock.Sig)
	}
}