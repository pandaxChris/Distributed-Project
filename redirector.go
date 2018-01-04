package main
/*
	The purpose of this package is to manage and redirect front end connections
	to a backend server. There will be 2 servers running and this will ping both
	servers to check which one is less busy
*/
import(
	"net"
	"fmt"
	"os"
	"strconv"
	"bufio"
	"time"
)

func main(){
	var serverOne, serverTwo int
	myChan := make(chan int, 2)
	listen, err := net.Listen("tcp", ":16000")
	if err != nil{
		fmt.Fprintln(os.Stderr, "Failed to startup on port 16000")
	}
	fmt.Fprintln(os.Stderr, "Started! Waiting for connections...")
	go func(){
		for{
			serverConnOne, err := net.Dial("tcp", "localhost:9000")
			if(err != nil){
				fmt.Fprintln(os.Stderr, "Failed to connect to server one")
			}
			serverConnOne.Write([]byte("NUMCONN\n"))
			myScanOne := bufio.NewScanner(serverConnOne)
			for myScanOne.Scan(){
				line := myScanOne.Text()
				line = string(line)
				myNum, err := strconv.Atoi(line)
				if err != nil{
					myChan <- -1

				}else {
					myChan <- myNum
				}
				break
			}
			serverConnOne.Close()
			serverConnTwo, err := net.Dial("tcp", "localhost:9001")
			if(err != nil){
				fmt.Fprintln(os.Stderr, "Failed to connect to server one")
			}
			serverConnTwo.Write([]byte("NUMCONN\n"))
			myScan := bufio.NewScanner(serverConnTwo)
			for myScan.Scan(){
				line := myScan.Text()
				line = string(line)
				myNum, err := strconv.Atoi(line)
				if err != nil{
					myChan <- -1
				}else{
					myChan <- myNum
				}
				break
			}
			serverConnTwo.Close()
			time.Sleep(2*time.Second)
		}
	}()

	for{

		conn, er := listen.Accept()
		if er != nil{
			fmt.Fprintln(os.Stderr, "Failed to startup on port 16000")
		}
		go func(){
			scan := bufio.NewScanner(conn)
			for scan.Scan(){
				line := scan.Text()
				if(line == "NUM"){
					serverOne = <- myChan
					serverTwo = <- myChan
				}
				if serverOne <= serverTwo || serverOne != -1{
					conn.Write([]byte("localhost:9000\n"))
				}else if serverTwo != -1{
					conn.Write([]byte("localhost:9001\n"))
				}else{
					conn.Write([]byte("FAIL"))
				}
				break
			}
			conn.Close()
		}()
	}
}