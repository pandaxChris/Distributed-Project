package main
/*
Chris Li
cl3250
CS3254: Parallel and Distributed Systems
Front End Handler
*/

import(
	"io/ioutil"
	"fmt"
	"net/http"
	"time"
	"html/template"
	"os"
	"net"
	"bufio"
	"strings"
)

//Make a connection to the server and return the client object
func makeConn() net.Conn{
	var line string= ""
	conn, err := net.Dial("tcp", "localhost:16000")
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Failed to connect to server")
		return nil
	}
	scan := bufio.NewScanner(conn)
	for scan.Scan(){
		line = scan.Text()
		break
	}
	if(line == "FAIL"){
		return nil
	}
	conn.Close()
	conn, err = net.Dial("tcp", line)
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Failed to connect to server")
		return nil
	}
	fmt.Fprintln("connected to: " + line)
	return conn
}
//Checks if a cookie is set
func cookieSet(name string, r *http.Request) bool{
	cookie, err := r.Cookie("parallel_user")
	//Return no cookie if an error arises
	if(err != nil){	return false }
	return cookie != nil && cookie.Value != ""
}


//Creates a new user if the user has not been named before.
func register(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodPost:	
		r.ParseForm()
		//Check empty form
		if(r.PostFormValue("uname") == "" || r.PostFormValue("pwd") == "" ||
			r.PostFormValue("name") == "" || r.PostFormValue("email") == ""){
			fmt.Fprintf(w, "<div>Invalid form data.</div><br/>"+getPageContents("signup"))
			break;
		}
		uname := r.PostFormValue("uname")
		pwd := r.PostFormValue("pwd")
		name := r.PostFormValue("name")
		email := r.PostFormValue("email")
		//Make connection to server
		conn := makeConn()
		if(conn == nil){
			fmt.Fprintf(w, "<div> Failed to Register</div><br/>"+getPageContents("signup"))
			break
		}else{
			//Follow Register protocol to the server
			stringToSend := "REGISTER," + uname + "," + pwd + "," + 
									name + "," + email
			conn.Write([]byte(stringToSend + "\r\n"))
			//Wait for a reply from the server and handle accordingly
			scan := bufio.NewScanner(conn)
			for scan.Scan(){
				text := scan.Text()
				fmt.Fprintln(os.Stderr, text)
				if(text == "FAIL"){	//Fail from server side
					fmt.Fprintf(w, "<div> Failed to Register</div><br/>"+getPageContents("signup"))
					break
				}else{
					fmt.Fprintln(os.Stderr, uname + " has successfully registered")
					http.Redirect(w,r, "home", http.StatusFound)
					break
				}
			}
		}
		conn.Close()
	case http.MethodGet:	//An on load
		cookie := cookieSet("parallel_user", r)
		if cookie {				//Make sure user not logged in and no error occurred
			fmt.Fprintf(w, "<script>window.location = 'home';</script>")
		}
		//fmt.Fprintf(w, "<script>window.location = 'home'</script>")
		fmt.Fprintf(w, getPageContents("signup"))
	}
}
//Sets up cookie for user login
func login(w http.ResponseWriter, r *http.Request){
	cookie := cookieSet("parallel_user", r)
	if(cookie){
		http.Redirect(w,r, "home", http.StatusFound)
		return
	}
	switch r.Method{
	//Only going to login if it's sent as a POST request
	case http.MethodPost:	
		r.ParseForm()
		uname := r.PostFormValue("uname")
		pwd := r.PostFormValue("pwd")
		if(uname == "" || pwd == ""){
			fmt.Fprintf(w, "<div>Failed to connect</div><br/>" +
				 getPageContents("login"))
			break
		}
		conn := makeConn()
		if conn == nil{
			fmt.Fprintf(w, "<div>Failed to connect</div><br/>" +
				 getPageContents("login"))
			fmt.Fprintln(os.Stderr, "Failed to login to user: " + uname)
		}else{
			//Send user credentials to the server
			conn.Write([]byte("LOGIN," + uname + "," +pwd + "\r\n"))
			conn.Write([]byte("\r\n"))
			//Get a reply from the server
			scan := bufio.NewScanner(conn)
			for scan.Scan(){
				text := scan.Text()
				fmt.Fprintln(os.Stderr, text)
				if(text == "FAIL") {
					fmt.Fprintf(w, "<div>Failed to login</div><br/>" +
						 getPageContents("login"))
					fmt.Fprintln(os.Stderr, "Failed to login to user:" + uname)
					break
				}else{
					cookie := http.Cookie{
						Name: "parallel_user",
						Value:  uname,
						Expires: time.Now().Add(60 * time.Minute)}
					http.SetCookie(w, &cookie)
					fmt.Println(os.Stderr, cookie.Value + " has logged in.")
					http.Redirect(w, r, "home", http.StatusTemporaryRedirect)
					break
				}
			}
			conn.Close()
		}
	case http.MethodGet:
		//Load the login page if user landed on login page without logging in
		fmt.Fprintf(w, getPageContents("login"))
	}
}	
//Get the contents of a file and return it
func getPageContents(page string) string{
	html, err  := ioutil.ReadFile(page + ".html")
	//Check if page is non-reachable and output to stdout if it isn't
	if(err != nil){
		//Return empty string on failure to open this file
		return ""
	}
	return string(html)
}

//	Handler for loading initial page(root url) and also to relocate if user tries to access
// 	a page that is not reachable
func loadPage(w http.ResponseWriter, r *http.Request){
	//Get url path for redirection
	url := r.URL.Path[1:] //Ignore first backslash so starts at index 1
	// Check if the file is a reachable destination
	page := getPageContents(url)
	//Send the user back to home page if the user tries to 
	//access a page that doesn't exist or access root
	if page == ""{	
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)	
	}else{
		fmt.Fprintf(w, string(page) )	//Load the page 
	}
}

//Handles everything to be done at home
func homeHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	default:
		//Load template
		temp := template.Must(template.ParseFiles("home.html"))
		//Pick what to show in navigation bar
		cookie, err := r.Cookie("parallel_user")
		//Surpress no cookie error -> taken care of in else
		if(err != nil) { }
		if(cookie == nil){
			temp.Execute(w, template.HTML("<nav><a href = '/home'>Home</a>" + 
							"<a href = '/login'>Login</a>" +
							"<a href= '/signup'>Sign Up </a></nav><br />"))
			return
		}
		if cookie != nil && cookie.Value != ""{
			conn := makeConn()
			if(conn == nil){
				temp.Execute(w, template.HTML("<nav><a href = '/home'>Home</a>" + 
								"<a href = '/login'>Login</a>" +
								"<a href= '/signup'>Sign Up </a></nav><br />"+
								"<span>Server unavailable at this moment</span>"))
				return
			}
			conn.Write([]byte("RETRP," + cookie.Value + "\r\n"))
			//Show up to first 50 posts that come up
			posts := make([]string, 0, 50)
			scan := bufio.NewScanner(conn)
			for scan.Scan(){
				text := scan.Text()
				posts = strings.Split(text, "+")
				break
			}
			conn.Close()
			//Navigation bar
			links := "<nav><a href = '/home'>Home</a>" + 
						"<a href = '/follow'>Follow People </a> "+
						"<a href = '/unfollow'>Unfollow people</a>"+
						"<a href='/post'>Make a Post </a> " +
						"<a href='/delete'> Delete my account </a>"+
						"<a href = '/logout'>Logout</a></nav><br /><br />"

			//Get the posts from all the users this user has followed
			if(len(posts) == 0){
				links += "<span> <i> No posts right now</i></span>"
			}else{
				for i := 0; i < len(posts)-1; i++{
					links += "<span stlye = ''>"+ posts[i] +"</span><br/><br/>"
				}
			}
			temp.Execute(w, template.HTML(links))
		}else{
			temp.Execute(w, template.HTML("<nav><a href = '/home'>Home</a>" + 
									"<a href = '/login'>Login</a>" +
									"<a href= '/signup'>Sign Up </a></nav>"))
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodPost:
		cookie, err := r.Cookie("parallel_user")
		if(err != nil){
			//Alert server an error has occurred with posting
			fmt.Println("An error occurred while posting")
			//If no cookie sent, send GET request to this page 
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)

		}
		r.ParseForm()
		conn := makeConn()
		conn.Write([]byte("POS,"+cookie.Value+","+r.PostFormValue("post") + "\r\n"))
		scan := bufio.NewScanner(conn)
		for scan.Scan(){
			text := scan.Text()
			fmt.Fprintln(os.Stderr, text)
			if(text == "DONE"){
				conn.Close()
				fmt.Fprintln(os.Stderr, cookie.Value + " Has posted: " + 
					r.PostFormValue("post"))
				break
			}else{
				conn.Close()
				fmt.Fprintln(os.Stderr, "Failed to process the post by "+ 
						cookie.Value)
				break
			}
			
		}
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	case http.MethodGet:
		fmt.Fprintf(w, getPageContents("post"))
	}
}
//Handles delete account
func deleteHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	case http.MethodPost:
		r.ParseForm()
		if( r.PostFormValue("contact") == "N" ){
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		}
		cook,err := r.Cookie("parallel_user")
		//Make sure there was a cookie in place before accessing this page
		if(err != nil){
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		}

		//Get the user's username
		name := cook.Value
		//Perform action to server
		conn := makeConn()
		conn.Write([]byte("DELETE," + name + "\r\n"))
		scan := bufio.NewScanner(conn)
		for scan.Scan(){
			text := scan.Text()
			fmt.Fprintln(os.Stderr, text)
			if(text == "DONE"){
				cook.MaxAge = -100
				//"Delete" cookie
				http.SetCookie(w, cook)
				//Log on front-end server
				fmt.Fprintln(os.Stderr, name + " has deleted their account")
				http.Redirect(w,r, "/home", http.StatusTemporaryRedirect)
				break
			}else{
				fmt.Fprintf(w, getPageContents("delete"))
				break
			}
		}
		conn.Close()
	case http.MethodGet:
		fmt.Fprintf(w, getPageContents("delete"))
	}
}
//Handler for the follow page
func followHandler(w http.ResponseWriter, r *http.Request){
	//Get cookie and then get the user associated with the cookie
	cookie, err := r.Cookie("parallel_user")
	if(err != nil){fmt.Println(err)}
	user := cookie.Value

	switch r.Method{
	case http.MethodGet:
		temp := template.Must(template.ParseFiles("follow.html"))
		//Get users that this user has not followed
		conn := makeConn()
		if conn == nil{
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
			return
		}
		conn.Write([]byte("GETUF," + user + "\r\n"))

		scan:= bufio.NewScanner(conn)
		unfollowed := make([]string, 0, 100)
		for scan.Scan(){
			text := scan.Text()
			fmt.Fprintln(os.Stderr, text)
			//Sent from server in form of follwo&unfollow
			twoArrays := strings.Split(text, "&")
			unfollowed = strings.Split(twoArrays[1], ",")
			break
		}
		conn.Close()
		outputStr := ""
		genString := ""
		if(len(unfollowed) == 1){
			genString += "<div> There are no available users to follow</div>"
		}else{
			//Go up to -1 to ignore DONE message from server
			for i:=0; i< len(unfollowed)-1; i++{
				genString += "<div id = 'usr'>"+
				" <form method = 'post' action = 'follow'>"+unfollowed[i]+" <br />"+
				"<input name = 'follow' value = " +unfollowed[i]+" type = 'hidden'>"+
				"<br /><button type = 'submit'> Follow</button></form></div>"

			}
		}
		outputStr += genString
		temp.Execute(w, template.HTML(outputStr))
	case http.MethodPost:
		//Assume only way here is after hitting the follow button
		//Does not consider people that try to reach here through
		//	direct modification of the packets being sent
		r.ParseForm()
		conn := makeConn()
		conn.Write([]byte("FOLLOW,"+ user + "," + r.PostFormValue("follow") +
		  			"\r\n"))
		scan := bufio.NewScanner(conn)
		for scan.Scan(){
			text := scan.Text()
			if(text == "DONE"){
				http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
				break
			}else{
				http.Redirect(w, r, "/follow", http.StatusResetContent)
				break
			}
		}
		conn.Close()
	}
}
//Handler for the unfollow page
func unfollowHandler(w http.ResponseWriter, r *http.Request){
	//Get cookie and then get the user associated with the cookie
	cookie, err := r.Cookie("parallel_user")
	if(err != nil){fmt.Fprintln(os.Stderr, err)}
	user := cookie.Value

	switch r.Method{
	case http.MethodGet:
		temp := template.Must(template.ParseFiles("unfollow.html"))
		//Get users that this user has followed
		conn := makeConn()
		if conn == nil{
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		}
		conn.Write([]byte("GETUF,"+user+ "\r\n"))
		scan:= bufio.NewScanner(conn)
		followed := make([]string,0,100)
		for scan.Scan(){
			text := scan.Text()
			fmt.Fprintln(os.Stderr, text)
			//Sent from server in form of follwo&unfollow
			twoArrays := strings.Split(text, "&")
			followed = strings.Split(twoArrays[0], ",")
			fmt.Fprintln(os.Stderr, followed)
			break
		}
		conn.Close()
		outputStr := ""
		if(len(followed) == 0 || len(followed) == 1){
			outputStr += "<div> You have not followed any users yet</div>"
		}else{
			for i:=0; i< len(followed); i++{
				outputStr += "<div id = 'usr'>"+
				"<form method = 'post' action = 'unfollow'>"+followed[i]+" <br />"+
				"<input name = 'unfollow' value = " +followed[i]+" type = 'hidden'>"+ 
				"<br /><button type = 'submit'> Unfollow</button></form></div>"
				
			}
		}
		temp.Execute(w, template.HTML(outputStr))
	case http.MethodPost:
		//Assume only way here is after hitting the follow button
		//Does not consider people that try to reach here through
		//	direct modification of the packets being sent
		r.ParseForm()
		conn := makeConn()
		conn.Write([]byte("UNFOLLOW,"+ user + "," + 
			r.PostFormValue("unfollow") + "\r\n"))
		scan := bufio.NewScanner(conn)
		for scan.Scan(){
			text := scan.Text()
			if(text == "DONE"){
				http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
			}else{
				http.Redirect(w, r, "/unfollow", http.StatusResetContent)
			}
			break
		}
		conn.Close()
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
	}
}
//Logs out the user by invalidating the cookie
func logOut(w http.ResponseWriter, r *http.Request){
	switch r.Method{
	default:
		cookie, err := r.Cookie("parallel_user")
		if err != nil{
			http.Redirect(w, r, "parallel_user", http.StatusTemporaryRedirect)
		}
		cookie.MaxAge = -100
		http.SetCookie(w, cookie)
		http.Redirect(w,r,"/home", http.StatusTemporaryRedirect)
	}
}
func main(){
	//Show server has started 
	fmt.Println("Server started")
	http.HandleFunc("/", loadPage)
	http.HandleFunc("/signup", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/follow", followHandler)
	http.HandleFunc("/logout", logOut)
	http.HandleFunc("/unfollow", unfollowHandler)
	//Listen on port 8080 for a connection
	err := http.ListenAndServe(":8080",  nil)
	//Alert if server failed to listen on 8080
	if err != nil{
		fmt.Println("Server start failed")
	}
}