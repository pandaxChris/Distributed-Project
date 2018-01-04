/*
Chris Li
cl3250
CS3254: Parallel and Distributed Systems
Replication Manager for server file consistancy
*/

package main 

import(
	"os"
	"net"
	"bufio"
	"fmt"
	"io/ioutil"
	//"crypto/md5"
	"sync"
	"strings"
)

//Mutexes created solely for updating files and arrays
var userLock sync.Mutex 
var postsLock sync.Mutex 
var followLock sync.Mutex

//Array to hold all servers that connect to this
// Assume that everyone who connects to this manager
// Is themselves a server
var connectedServers []Server
type Server struct{
	conn net.Conn 
	address string
}

/*
	Description: Check to see if a server is connected 
	Output: Adds a server to the array if it isn't already there
*/
func checkInArray(remoteAddr string, connection net.Conn){
	//First check to see if the server is already in the list of servers
	for _, serv := range connectedServers{
		if serv.address == remoteAddr{
			return 
		}
	}
	connectedServers = append(connectedServers, Server{connection, remoteAddr})
}

/*	
	Description: Deletes a given user in the system
	Returns: True or False based on success of deletion
*/
func deleteUser(uname string) bool{
	userLock.Lock()
	postsLock.Lock()
	followLock.Lock()
	defer userLock.Unlock()
	defer postsLock.Unlock()
	defer followLock.Unlock()
	file, err := ioutil.ReadFile("users_main.txt")
	users := make([]string, 0, 100)
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Failed to open user file")
		return false
	}

	fileString := string(file)
	fileContents := strings.Split(fileString, "\n")	
	//Get contents of the file and then remove the user
	for i := 0; i < len(fileContents)-1; i++{
		info := strings.Split(fileContents[i], ",")
		//Add the user to the list of all users
		users = append(users, info[0])
		//Remove the user if the username matches
		if(info[0] == uname){
			fileContents[i] = ""
		}else{	//Else add a newline character to the end
			fileContents[i]+= "\n"
		}
	}

	//Overwrite the existing file without inclusion of deleted user
	output := strings.Join(fileContents, "")
	errUser := ioutil.WriteFile("users_main.txt", []byte(output), 0644)

	errPost := errUser
	//Delete all their posts
	filePost, errPost := ioutil.ReadFile("posts_main.txt")
	if errPost == nil{
		fileStringPost := string(filePost)
		fileContentsPost := strings.Split(fileStringPost, "\n")
		for i := 0; i < len(fileContentsPost)-1; i++{
			post := strings.Split(fileContentsPost[i], ":")
			if(post[0] == uname){	//Delete if owner matches
				fileContentsPost[i] = ""
			}else{
				fileContentsPost[i]+= "\n"
			}
		}
		outPost := strings.Join(fileContentsPost, "")
		errPost = ioutil.WriteFile("posts_main.txt", []byte(outPost), 0644)
	}

	//Delete all the relationships that this user follows
	errFollow := false
	fileFol, err := ioutil.ReadFile("following_main.txt")
	followingString := string(fileFol)
	followingArray := strings.Split(followingString, "\n")
	fmt.Fprintln(os.Stderr, followingArray)
	for i := 0; i < len(followingArray)-1; i++{
		currentLine := strings.Split(followingArray[i], "->")
		if (currentLine[0] == uname || currentLine[1] == uname){
			followingArray[i] = ""
		}else{
			followingArray[i] += "\n"
		}
	}
	outFollows := strings.Join(followingArray, "")
	er := ioutil.WriteFile("following_main.txt", []byte(outFollows), 0644)
	if er != nil{
		errFollow = true
	}

	return errUser == nil && errPost == nil && !errFollow
}
/*
	Description: Adds a user post to the system
	Returns: Boolean based on success of addition to file
*/
func addPost(uname string, post string) bool{
	postsLock.Lock()
	defer postsLock.Unlock()
	file, err := ioutil.ReadFile("posts_main.txt")
	if(err != nil){
		fmt.Fprintln(os.Stderr,"Failed to open post file")
		return false
	}
	//Cast slice of entire byte array to a string
	fileString := string(file)
	//Split on the : delimeter 
	fileString += uname + ":" + post + "\n"
	//Write back the file the contents in addition the post
	err = ioutil.WriteFile("posts_main.txt", []byte(fileString), 0644)
	return err == nil
}
/*
	Description: Remove user following pair to the system(unidirectional follow)
	Returns: Boolean on success of deletion
*/
func unfollowUser(userOne, userTwo string) bool{
	followLock.Lock()
	defer followLock.Unlock()
	file, err := ioutil.ReadFile("following_main.txt")
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Failed to open following file")
		return false
	}
	//Check to first make sure the file is not empty
	if(len(file) == 0){
		return false
	}
	fileString := string(file)
	//Get each following pair
	follows := strings.Split(fileString, "\n")
	outputString := ""
	//Go through the file to find the follow pair
	for i:= 0; i < len(follows)-1; i++{
		follower := strings.Split(follows[i], "->")
		//Check if the user already followed this user
		if follower[0] == userOne && follower[1] == userTwo{
			continue
		}else{
			outputString += follows[i] +"\n"
		}
	}
	err = ioutil.WriteFile("following_main.txt", []byte(outputString), 0644)
	return err == nil

}
/*
	Description: Add user following pair to the system(unidirectional follow)
	Returns: Boolean on success of addition
*/
func followUser(userOne, userTwo string) bool{
	followLock.Lock()
	defer followLock.Unlock()
	file, err := ioutil.ReadFile("following_main.txt")
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Failed to open following file")
		return false
	}
	fileString := string(file)
	//Get each following pair
	follows := strings.Split(fileString, "\n")
	//This won't execute if the file is empty
	for i:= 0; i < len(follows)-1; i++{
		follower := strings.Split(follows[i], "->")
		//Check if the user already followed this user
		if follower[0] == userOne && follower[1] == userTwo{
			return false
		}
	}
	fileString += userOne + "->" + userTwo + "\n"
	err = ioutil.WriteFile("following_main.txt", []byte(fileString), 0644)
	return err == nil
}

/*
	Description: Add user credentials to the file of users
	Returns: bool based on successful creation(true) or not(false)
*/
func createUser(uname string, pass string, name string, email string) string{
	userLock.Lock()
	//Get the file contents of the users file into a byte array
	file, err := ioutil.ReadFile("users_main.txt")
	//Cast slice of entire byte array to a string
	fileString := string(file)
	//Check if user already exists
	fileContents := strings.Split(fileString, "\n")	
	for i:= 0; i < len(fileContents); i++{
		//User information will be in "info" in form of [user,pass,name,email]
		info := strings.Split(fileContents[i], ",")
		if(uname == info[0]){
			return "FAIL"
		}
	}
	//Add user to the file list
	newLine := uname + "," + pass + "," + name + "," + email + "\n"
	newFile :=  fileString + newLine
	err = ioutil.WriteFile("users_main.txt", []byte(newFile), 0644)
	if err != nil{
		return "FAIL"
	}
	//Automatically follow self
	followUser(uname, uname)
	defer userLock.Unlock()
	return "DONE"
}
func main(){
	//We will have one global copy of the files 
	connectedServers := make([]Server, 0, 10)
	
	// denoted with the _main suffix
	_, userStatus := os.Stat("users_main.txt")
	if(os.IsNotExist(userStatus)){
		user, _ := os.Create("users_main.txt")
		user.Close()
	}

	_, postsStatus := os.Stat("posts_main.txt")
	if(os.IsNotExist(postsStatus)){
		posts, _ := os.Create("posts_main.txt")
		posts.Close()
	}

	_, followStatus := os.Stat("following_main.txt")
	if(os.IsNotExist(followStatus)){
		following, _ := os.Create("following_main.txt")
		following.Close()
	}
	//Setup port to lsiten for requests
	listen, err := net.Listen("tcp", "localhost:15000")
	if err != nil{
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintln(os.Stderr, "Replication Manager Started")
	for{
		//Listen for connections
		conn, err := listen.Accept()
		if(err != nil){
			fmt.Fprintln(os.Stderr,"Something went wrong on accepting client")
			continue
		}
		go func(){
			connAddr := conn.RemoteAddr().String()
			fmt.Fprintln(os.Stderr, connAddr + " has connected.") 
			//Check in array and if not there then add it
			checkInArray(conn.RemoteAddr().String(), conn)
			scanner := bufio.NewScanner(conn)
			for scanner.Scan(){
				line := scanner.Text()
				fmt.Println(os.Stderr, "Got:" + line)
				// Command: GRAB 
				// Purpose: Retrieve current version of server files 
				if line == "GRAB"{
					//Lock files to ensure no race conditions
					// Defer unlocks till the end to ensure consistancy
					userLock.Lock()
					followLock.Lock()
					postsLock.Lock()

					userItems, _ := ioutil.ReadFile("users_main.txt")
					followItems, _ := ioutil.ReadFile("following_main.txt")
					postItems, _  := ioutil.ReadFile("posts_main.txt")

					//Write file items to the server asking for it
					conn.Write([]byte(string(userItems) + "\r\n"))
					conn.Write([]byte("NEXT\r\n"))
					conn.Write([]byte(string(followItems) +"\r\n"))
					conn.Write([]byte("NEXT\r\n"))
					conn.Write([]byte(string(postItems) + "\r\n"))

					//Unlock after sent to server
					userLock.Unlock()
					followLock.Unlock()
					postsLock.Unlock()
					//Close the connection
					conn.Close()
				}else{
					params := strings.Split(line, ",")
					fmt.Fprintln(os.Stderr,params)
					// Params[0] holds operation
					// Params[1] holds fileName
					// Params[2] holds string separated by ;
					switch params[0]{
					case "ADD":
						switch params[1]{
						case "user":
							myItems := strings.Split(params[2], ";")
							createUser(myItems[0], myItems[1], myItems[2], myItems[3])
						case "follow":
							myItems := strings.Split(params[2], ";")
							followUser(myItems[0], myItems[1])
						case "post":
							myItems := strings.Split(params[2], ";")
							addPost(myItems[0], myItems[1])
						}
					case "DELETE":
						switch params[1]{
						case "user":
							deleteUser(params[2])
						case "follow":
							myItems := strings.Split(params[2], ";")
							unfollowUser(myItems[0],myItems[1])
						}
					}

					//Updates everyone else's files
					// Server that pushed change will update after on own
					// if they were able to push change in first place
					for i := 0; i < len(connectedServers); i++{
						fmt.Fprintln(os.Stderr, "Got here")
						if(conn.RemoteAddr().String() == connectedServers[i].address){
							continue
						}
						fmt.Fprintln(os.Stderr, "Sending to "+connectedServers[i].address)
						serverConn, serr := net.Dial("tcp", connectedServers[i].address)
						if(serr != nil){
							fmt.Fprintln(os.Stderr, "Failed to write to server:" +
									connectedServers[i].address)
						}
						serverConn.Write([]byte(line + "\n"))
						serverConn.Close()
						fmt.Fprintln(os.Stderr, "Wrote to server:" + 
							connectedServers[i].address)
					}

				}
			}
		}()

	}
}