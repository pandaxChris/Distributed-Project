/*
Chris Li
cl3250
CS3254 - Parallel and Distributed Systems
Back end server
*/

package main

import(
	"io/ioutil"
	"net"
	"bufio"
	"fmt"
	"strings"
	"os"
	"sync"
)
// Three locks for locking files
var postsLock sync.Mutex
var followLock sync.Mutex
var userLock sync.Mutex
var numLock sync.Mutex
/*
	Description: Get all the followed and unfollowed users
	Returns: String array of followed users; string array of non-followed users
*/
func getFollowedUsers(user string) ([]string, []string){
	//get all users
	users := make([]string, 0, 100)
	userLock.Lock()
	userFile, _ := ioutil.ReadFile("users.txt")
	userSplit := strings.Split(string(userFile), "\n")
	for i := 0; i < len(userSplit); i++{
		user := strings.Split(userSplit[i], ",")
		users = append(users, user[0])
	}
	userLock.Unlock()
	followLock.Lock()
	followed := make([]string, 0, 100)
	unfollowed := make([]string, 0, 100)
	followFile, _ := ioutil.ReadFile("following.txt")
	followSplit := strings.Split(string(followFile), "\n")
	//First get all the users that this user has followed
	for i := 0; i < len(followSplit); i++{
		pair := strings.Split(followSplit[i], "->")
		if(pair[0] == user){
			followed = append(followed, pair[1])
		}
	}
	//Next get the users that this user isn't following
	for i := 0; i < len(users); i++{
		isFollower := true
		for j := 0; j < len(followed); j++{
			if followed[j] == users[i]{
				isFollower = true
				break
			}
			isFollower = false
		}
		if !isFollower{
			unfollowed = append(unfollowed, users[i])
			isFollower = true
		}
	}
	defer followLock.Unlock()
	//Return both arrays
	return followed,unfollowed 
}
/*
	Description: Get all the posts that are visible to this user
	Returns: String array of all posts this user can see
*/
func getVisiblePosts(user string) []string{
	//First get all the users this user is following
	followed := make([]string, 0, 100)
	followLock.Lock()
	postsLock.Lock()
	
	followFile, _ := ioutil.ReadFile("following.txt")
	followSet := strings.Split(string(followFile), "\n")
	for i := 0; i < len(followSet); i++{
		users := strings.Split(followSet[i], "->")
		if(users[0] == user){
			followed = append(followed, users[1])
		}
	}
	//Get all posts from those that this user has followed
	posts := make([]string, 0, 100)
	postsFile, _ := ioutil.ReadFile("posts.txt")
	if(len(postsFile) == 0){
		return posts
	}
	postsSplit := strings.Split( string(postsFile), "\n")
	for i := 0; i < len(postsSplit); i++ {
		for j := 0; j < len(followed); j++{
			post := strings.Split(postsSplit[i], ":")
			if post[0] == followed[j]{
				posts = append(posts, postsSplit[i])
			}
		}
	}
	defer postsLock.Unlock()
	defer followLock.Unlock()
	return posts
}
/*
	Description: Remove user following pair to the system(unidirectional follow)
	Returns: Boolean on success of deletion
*/
func unfollowUser(userOne, userTwo string) bool{
	followLock.Lock()
	defer followLock.Unlock()
	file, err := ioutil.ReadFile("following.txt")
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
	err = ioutil.WriteFile("following.txt", []byte(outputString), 0644)

	serverConnect , error := net.Dial("tcp", "localhost:15000")
	if(error != nil){
		fmt.Fprintln(os.Stderr, "Failed to alert other servers")
	}
	serverConnect.Write([]byte("DELETE,follow," + userOne + ";" + userTwo +"\n"))
	serverConnect.Close()
	return err == nil

}
/*
	Description: Add user following pair to the system(unidirectional follow)
	Returns: Boolean on success of addition
*/
func followUser(userOne, userTwo string) bool{
	followLock.Lock()
	defer followLock.Unlock()
	file, err := ioutil.ReadFile("following.txt")
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
	err = ioutil.WriteFile("following.txt", []byte(fileString), 0644)

	serverConnect , error := net.Dial("tcp", "localhost:15000")
	if(error != nil){
		fmt.Fprintln(os.Stderr, "Failed to alert other servers")
	}
	serverConnect.Write([]byte("ADD,follow," + userOne + ";" + userTwo +"\n"))
	serverConnect.Close()
	return err == nil
}
/*
	Description: Adds a user post to the system
	Returns: Boolean based on success of addition to file
*/
func addPost(uname string, post string) bool{
	postsLock.Lock()
	defer postsLock.Unlock()
	file, err := ioutil.ReadFile("posts.txt")
	if(err != nil){
		fmt.Fprintln(os.Stderr,"Failed to open post file")
		return false
	}
	//Cast slice of entire byte array to a string
	fileString := string(file)
	//Split on the : delimeter 
	fileString += uname + ":" + post + "\n"
	//Write back the file the contents in addition the post
	err = ioutil.WriteFile("posts.txt", []byte(fileString), 0644)
	serverConnect , error := net.Dial("tcp", "localhost:15000")
	if(error != nil){
		fmt.Fprintln(os.Stderr, "Failed to alert other servers")
	}
	serverConnect.Write([]byte("ADD,post," + uname + ";" + post + "\n"))
	serverConnect.Close()
	return err == nil
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
	file, err := ioutil.ReadFile("users.txt")
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
	errUser := ioutil.WriteFile("users.txt", []byte(output), 0644)

	errPost := errUser
	//Delete all their posts
	filePost, errPost := ioutil.ReadFile("posts.txt")
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
		errPost = ioutil.WriteFile("posts.txt", []byte(outPost), 0644)
	}

	//Delete all the relationships that this user follows
	errFollow := false
	fileFol, err := ioutil.ReadFile("following.txt")
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
	er := ioutil.WriteFile("following.txt", []byte(outFollows), 0644)
	if er != nil{
		errFollow = true
	}
	serverConnect , error := net.Dial("tcp", "localhost:15000")
	if(error != nil){
		fmt.Fprintln(os.Stderr, "Failed to alert other servers")
	}
	serverConnect.Write([]byte("DELETE,user," + uname + "\n"))
	serverConnect.Close()
	return errUser == nil && errPost == nil && !errFollow
}
/*
	Description: Logs in the user
	Returns: If valid user, returns string with user credentials separated by ,
			If invalid user, returns empty string
*/
func loginUser(uname string, pass string) string{
	file, err := ioutil.ReadFile("users.txt")
	if(err != nil){
		fmt.Fprintln(os.Stderr,"Failed to open user file")
	}
	fileString := string(file)
	fileContents := strings.Split(fileString, "\n")	
	//Find the user in the system - won't run if the file is empty
	for i:= 0; i < len(fileContents); i++{
		//User information will be in "info" in form of [user,pass,name,email]
		info := strings.Split(fileContents[i], ",")
		if(uname == info[0]){
			if(pass == info[1]){	//Check to make sure passwords match
				return fileContents[i][:len(fileContents[i])]
			}else{ //Invalid login
				return "FAIL"
			}
		}
	}
	//No user found
	return "FAIL"
}
/*
	Description: Add user credentials to the file of users
	Returns: bool based on successful creation(true) or not(false)
*/
func createUser(uname string, pass string, name string, email string) string{
	userLock.Lock()
	//Get the file contents of the users file into a byte array
	file, err := ioutil.ReadFile("users.txt")
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
	err = ioutil.WriteFile("users.txt", []byte(newFile), 0644)
	if err != nil{
		return "FAIL"
	}
	//Automatically follow self
	followUser(uname, uname)
	defer userLock.Unlock()
	serverConnect , error := net.Dial("tcp", "localhost:15000")
	if(error != nil){
		fmt.Fprintln(os.Stderr, "Failed to alert other servers")
	}
	serverConnect.Write([]byte("ADD,user,i" + uname + ";" + pass + ";" +
							name + ";" + email+"\n"))
	serverConnect.Close()
	return "DONE"
}
/*
	Description: updates file 'fileName' with 'content' by donig 'how'
	Returns: Nothing
*/
func updateFile(fileName string, how string, content string){
	if how == "ADD"{
		params := strings.Split(content,";")
		switch fileName{
		case "user":
			createUser(params[0],params[1],params[2],params[3])
		case "follow":
			followUser(params[0],params[1])
		case "post":
			addPost(params[0],params[1])

		}
	}else if how == "DELETE"{
		params := strings.Split(content,",")
		switch fileName{
		case "user":
			deleteUser(content)
		case "follow":
			unfollowUser(params[0],params[1])
		}
	}
}
func main(){
	numConns := 0
	// Open on port 9000 and listen for connections
	listen, err := net.Listen("tcp", ":9001")
	if err != nil{
		fmt.Fprintln(os.Stderr, "Failed to listen on port 9000")
		return
	}
	defer listen.Close()
	
	//First get a fresh copy of a file from the replication manager
	conn, err := net.Dial("tcp", "localhost:15000")
	if err != nil{
		fmt.Fprintln(os.Stderr, "Failed to connect to replication manager")
		fmt.Fprintln(os.Stderr, "...Going local")
		_, userStatus := os.Stat("users.txt")
		if(os.IsNotExist(userStatus)){
			user, _ := os.Create("users.txt")
			user.Close()
		}

		_, postsStatus := os.Stat("posts.txt")
		if(os.IsNotExist(postsStatus)){
			posts, _ := os.Create("posts.txt")
			posts.Close()
		}

		_, followStatus := os.Stat("following.txt")
		if(os.IsNotExist(followStatus)){
			following, _ := os.Create("following.txt")
			following.Close()
		}
	}else{
		conn.Write([]byte("GRAB\n"))
		scanner := bufio.NewScanner(conn)
		userFile, _ := os.Create("users.txt")
		postsFile, _ := os.Create("posts.txt")
		followingFile, _ := os.Create("following.txt")
		currentIdx := 0
		myFiles := [3]*os.File{userFile, postsFile, followingFile}
		for scanner.Scan(){
			nextLine := scanner.Text()
			if(nextLine == "END"){
				break
			}
			if(nextLine == "NEXT"){
				currentIdx++
				continue
			}
			myFiles[currentIdx].Write([]byte(nextLine))

		}
		userFile.Close()
		followingFile.Close()
		postsFile.Close()
	}
	fmt.Fprintln(os.Stderr, "Server started")
	// We check to see if the resources we want exist
	// 		(users.txt, posts.txt, following.txt)
	// If they don't exist, create them

	/*
		Protocol to be setup:
		Will accept requests in the form of a string array
		arr[0] will contain what the action to perform is
		The list of actions are:
			CreateUser - REGISTER
			LoginUser - LOGIN
			DeleteUser - DELETE
			CreatePost - CREATE
			FollowU - FOLLOW
			UnfollowU - UNFOLLOW
			Post - POST
			getPosts - GETP
		The next indices will indicate params for the functions associated
			CreateU - uname, pwd, name, email
			LoginU - uname, pwd
			DeleteU - uname
			CreateP - uname, post
			FollowU - userOne, userTwo
			UnfollowU - userOne, userTwo
			PostU - user,post
			GETP - user
		Will send the string 'DONE' back to the user on successful operation

		Will send the string "FAIL" back to the user in the case any of these
		operations were to fail
	*/


	for{

		//Handle each connection as a thread so they can run in parallel
		//Open a connection to accept incoming connections
		conn, err := listen.Accept()
		if err != nil{
			fmt.Fprintln(os.Stderr, "Failed to accept connecting client")
			return
		}

		go func(){
			numLock.Lock()
			numConns++
			numLock.Unlock()
			//Create a buffer for reading data on the accepted connection
			scan := bufio.NewScanner(conn)
			// for the incoming conenction, check what they want
			// Log on the server the result of each if they are successful
			for scan.Scan(){	
				//Convert scanner input to a string
				line := scan.Text()
				fmt.Fprintln(os.Stderr, "Accepted: " + line)
				params := strings.Split(line,",")
				//fmt.Fprintln(os.Stderr, params)
				if params[0] == "REGISTER"{
					if len(params) < 3{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					uname := params[1]
					pwd := params[2]
					name := params[3]
					email := params[4]
					suc := createUser(uname, pwd, name, email)
					conn.Write([]byte(suc + "\r\n"))
					if(suc == "DONE"){
						fmt.Fprintln(os.Stderr, uname + " has registered.")
					}
				}else if params[0] == "LOGIN" {
					if len(params) < 3{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					userInfo := loginUser(params[1], params[2])
					conn.Write([]byte(userInfo + "\r\n"))
					//Alert server this user has logged in
					user := strings.Split(userInfo, ",")
					if userInfo != ""{
						fmt.Fprintln(os.Stderr, user[0] + " has logged in")
					}
				}else if params[0] == "DELETE"{
					if len(params) < 2{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					suc := deleteUser(params[1])
					if suc{
						conn.Write([]byte("DONE"+ "\r\n"))
						fmt.Fprintln(os.Stderr, params[1] + " has deleted their account.")
					}else{
						conn.Write([]byte("FAIL"+ "\r\n"))
					}
				}else if params[0] == "FOLLOW"{
					if len(params) < 3{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					suc := followUser(params[1], params[2])
					if(suc){
						conn.Write([]byte("DONE"+ "\r\n"))
						fmt.Fprintln(os.Stderr, params[1] + " has followed " + params[2])
					}else{
						conn.Write([]byte("FAIL"+ "\r\n"))
					}
				}else if params[0] == "UNFOLLOW"{
					if len(params) < 3{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					suc := unfollowUser(params[1], params[2])
					if(suc){
						conn.Write([]byte("DONE"+ "\r\n"))
						fmt.Fprintln(os.Stderr, params[1] + " has unfollowed " + params[2])
					}else{
						conn.Write([]byte("FAIL"+ "\r\n"))
					}
				}else if params[0] == "RETRP"{
					if len(params) < 2{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					posts := getVisiblePosts(params[1])
					str := ""
					for i := 0; i < len(posts); i++{
						str += posts[i] + "+"
					}
					conn.Write([]byte(str + "\r\n"))
				}else if params[0] == "POS"{
					if len(params) < 3{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					suc := addPost(params[1], params[2])
					if(suc){
						conn.Write([]byte("DONE" + "\r\n"))
						fmt.Fprintln(os.Stderr, params[1] + " posted " + params[2])
					}else{
						conn.Write([]byte("FAIL"+ "\r\n"))
					}
				}else if params[0] == "GETUF"{
					if len(params) < 2{
						conn.Write([]byte("FAIL" + "\r\n"))
						continue
					}
					fol, unfol := getFollowedUsers(params[1])
					conn.Write( []byte( strings.Join(fol,",") + "&" + 
						strings.Join(unfol, ",") +"DONE"+ "\r\n") )
				}else if params[0] == "CLOSE"{
					break
				}else if params[0] == "FILEUPDATE"{
					//Define the things to update
					//Params[1] will hold the filename
					// Params[2] will hold the line to override
					// Calls updateFile(name, how, content)
					// Note: Content utilizes ; instead of ,
					how := params[0]
					fileName := params[1]
					content := params[2]
					updateFile(fileName, how, content)
				}else if params[0] == "NUMCONN"{
					numLock.Lock()
					conn.Write([]byte(string(numConns) + "\r\n"))
					numLock.Unlock()
				}else{
					conn.Write([]byte("FAIL"+ "\r\n"))
					fmt.Fprintln(os.Stderr, "Someone has entered a wrong command")
				}
				
			}
			conn.Close()

		}()	//End closure func

	}
}