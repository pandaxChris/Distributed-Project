Chris Li
cl3250 
CS3254 Parallel and Distributed Systems

####################################
#	   [Part One]	           #
####################################
Usage:
The folder containing all the files runs under main.go. Once started, it'll listen on port 8080 for any incoming connections and will handle connections one at a time. 

What the website handles:

Redirection - Users can move from one page to another through the anchor tags on the page. If the user attempts to change the url to reach an unreachable file, it will redirect to the home page. 

Social Media - Users can register and log into unique user accounts. These accounts are stored in a global array(usr) for now. Users can also post messages that go up to 100 in length. Users can also destroy their account as well as log out of their accounts.



What may go wrong(Things that can break):

1. If the user keeps the cookie and the server restarts, any attempt to display other users’ stuffs will cause the server to panic as there are no current users allocated(not even self)

####################################
#	    [Part Two]		   #
####################################
The contents of this folder contains multiple HTML files that will help create
a front end website in which users can have basic social-media functionalities
such as:

Register
Login
Follow Users
Unfollow Users
Delete your account
Logout


There are also two .go files main.go and server.go which act as servers for
hosting the site(main.go) and a backend server(server.go) in which all the 
data for the website is stored. These two .go files communicate through
TCP/IP through a protocol created by myself. There is a basic syntactical
string format that follows this format:

STRING = COMMAND,ARG1,ARG2,...

This string is then sent to the server port and then be acted upon accordingly,
often returning a DONE or FAIL message back to the front end server.

The following commands and arguments used are:

REGISTER(USERNAME, PASSWORD, NAME, EMAIL) - Registers a user with these credentials
LOGIN(USERNAME,PASSWORD) - Logs a user in using these credentials
DELETE(USERNAME) - Deletes the user with this username
FOLLOW(USERONE, USERTWO) - Have USERONE follow USERTWO
UNFOLLOW(USERONE, USERTWO) - Have USERONE unfollow USERTWO
RETRP(USERNAME) - Retrieves all the posts visible by this user
POS(USERNAME,MESSAGE) - Post MESSAGE for the user
GETUF(USERNAME) - Get two lists(one of unfollowed users; one for followed)
CLOSE - Close the connection

Note: This protocol does not enforce any security and does not verify these
are valid messages(and not from some adversary telnet'ing with known accounts)

Note 2: If any user decides to telnet into the server and run a command with
more than the parameters specified, it will truncate any extra parameters given.

####################################
#	   [Part Three]            #
####################################
Files that are shared on the server: users.txt, files.txt, posts.txt, follow.txt, the jobs channel

Building on top of the second part of the project, this part handles concurrent connections and how jobs are handled. This is implemented through the jobs channel which is a channel of type "Job" with the fields conn and request. This channel is used in the performRequest method which takes the channel as a parameter and performs the next job in the channel. Users connecting to the server
will write to the channel instead of writing to the files themselves. This will allow for a FIFO ordering of the jobs that are given to the server and will give back a response to the client at that current time of where the job is on the channel. 

The performRequest function is taken care of in a goroutine, which may be a possible flaw in itself as the scheduler may delay the action to perform to a far later time. If this is the case, it may heavily  delay the service to the users. But the application does prioritize sending tasks to the channel and will do so in a FIFO order so the request will process eventually in the order that they arrive.

Revisions made: 
1. The job channel was taken out and locks were implemented to ensure safety of writing to shared resources. 

####################################
#	   [Part Four]		   #
####################################
Replication is achieved through a central replication manager. 
The processes should be started in this order: 
ReplicationManager.go -> Redirector.go -> ServerOne.go/ServerTwo.go -> Main.go

On first startup, servers will update their files(if they don’t already exist) before listening for connections. Once they’re done updating, they will listen for connections which in turn will run normally as they should in Part Three. In addition to part three, any changes made to a server file will also be pushed to the replication manager which handles incoming changes by changing its global files(denoted with the “_main” in the name.

Upon receiving an update on the replication manager, it will update its own files and then push the change to the other servers(the replication manager keeps a list of all the servers that ever connected to it - assumes that only servers are connecting to it)

Things to note:
	Ports used:
		Backend Servers: 9000-9001
		Client Server: 8080
		Replication Manager: 15000
		Redirecting Server: 16000 
What does not work:
Security -> logging in with same accounts

