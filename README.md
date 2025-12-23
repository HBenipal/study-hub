# Study Hub

## Links
- Video URL: https://www.youtube.com/watch?v=IocTEDh-h1w&feature=youtu.be

## Description

- This is a web-based collaborative Study Hub platform
- On this platform, users have the ability to create study rooms for collaborative studying
- Each room can have multiple documents, that can be edited collaboratively in real-time (similar to Google Docs)
- Additionally, users have the option to upload PDF files to a study room too
- The collaborative editor also has a "Talk to AI" feature. Users have the ability to speak a prompt to directly add content to their PDF. The AI integration is context-aware. For example, if the user is on a document with just a title "Essay on Javascript" and they ask the AI to "write the first paragraph", a paragraph related to Javascript will be directly inserted into the document, and this update will also be broadcasted to all other users currently on that document
- The markdown editor, as the name suggests, allows users to input markdown text, and a live preview is always visible on the right side of the screen. This allows users to keep track of what their formatted document looks like at all times.
- In addition to markdown, users also have the ability to include LaTeX content directly into their document. The AI integration is extended to be able to output both markdown and LaTeX directly into the document
- Each room must have a name, which doesn't have to be unique
- Each room also has a unique room code, through which users can access any room in the study hub
- Rooms can be public or private. Users have the ability to create both public and private rooms.
- Public rooms can be accessed directly from the home page of the app, which shows a paginated list of all public rooms available
- On the other hand, private rooms can only be accessed through their room code. Users are free to either directly enter the room URL into their browser, or enter the room code on the home page of the app in order to join rooms
- Users have the option to either sign up using a username and password, or using GitHub.
- Logging in with github has an additional advantage- it enabled users to upload PDF documents in rooms. Any document uploaded by a user is stored transparently on a GitHub repository that is automatically created on their account as soon as they sign up with or link their GitHub account

## Development

### Tech Stack
#### Frontend
- On the frontend, we're using NextJS with Typescript
- TailwindCSS is used for styling, [Lucide-React](https://lucide.dev) is used for Icons and [Shadcn-ui](https://ui.shadcn.com) is used for its tooltip and toast components
- For parsing markdown into HTML tags, the [Marked](https://github.com/markedjs/marked)
- [Katex](https://katex.org) is used for parsing LaTeX

#### Backend
- Our backend is written in Go
- For the most part, we use the native `net/http` for handling API requests
- [go-chi](https://github.com/go-chi/chi) is used for routing and parsing URL parameters. We chose this because of ease of adding middleware and grouping routes
- We also use [gorilla websockets](https://github.com/gorilla/websocket) for websockets, and [gorilla session] (https://github.com/gorilla/sessions) for managing sessions
- [lib/pq](https://github.com/lib/pq) is used for managing the Postgres, and [go-redis](https://github.com/redis/go-redis) is used for managing Redis
- Utility libraries: [crypto](https://pkg.go.dev/golang.org/x/crypto@v0.45.0) for bcrypt hashing. [google UUID](https://pkg.go.dev/github.com/google/uuid@v1.6.0) for generating unique user IDs [godotenv](https://pkg.go.dev/github.com/joho/godotenv@v1.5.1) for loading environment files while developing locally. [OpenAI go library](https://github.com/openai/openai-go) for accessing OpenAI APIs
#### Data
- Postgres is used as the SQL database for our application
- Redis is used for caching live document states

### Architecture
#### Realtime editing
- We achieve realtime editing by using websockets. When a user joins a room, they connect to the websocket endpoint and pass in their room code and the current document id. This allows us to store that specific user internally on the backend (in-memory)
- The main means of back and forth communication using websockets is `Messages`
- Each message has a non-optional field `Type`. We support the following message types:
   - `init`: This is the first message that is sent by the server to the user once the connect. This includes information such as the initial document content and the number of online users on that specific document
   - `clientCount`: This message is sent by the server to all users connected to a specific document in a specific room. Through this message, the frontend updates the live client count, which represents the number of users who currently have that specific document open
   - `documentListUpdate`: This message is sent by the server to all the users connected to that specific room, and it indicates that the document list has changed. Upon recieving this message, the frontend re-fetches the list of documents for the current room from the server. This ensures that the document list is always up to date, and users don't have to refresh the page in order to see newly added documents
   - `operation`: This is the heart of our realtime functionality. It will be described in more details below.
- When a user edits a document on the frontend, this is how the information flows:
   - The frontend computes a diff of the whatever content the user added in the last 100ms. Based on this diff, an operation in calculated.
   - An operation can be of two types: insert or delete. Insert operations must include the position where some text was inserted, and the content that was inserted. Delete operations include the the index where text was deleted and the length of the deleted string.
     - For example, if the document's contents are "He" and the user types "ello" the operation generated would have text "ello", type "insert" and position 2. Delete would work similarly
  - When this operation is calculated, from the client side a websocket message of type operation is sent, which includes the generated operation and the id of the user who generated this operation
  - Upon recieving this operation, the backend retrieves the current document state (if it is not already in redis, we fetch the current state from postgres and insert it to redis using keys of form `doc:{roomCode}:{documentId}:content`. If it is also not in postgres, we create the document and insert it in both postgres and redis.
  - Using the retrieved content and the recieved operation, the backend applies that operation onto the content to generate the updated document content. Note that this is the same content that the sender of this operation sees on their screen. But other users on the same document don't see this just yet. So, the server sends an message of type `operation` to all users connected to that room and that document.
  - Upon recieving that operation, other users apply that insert/delete onto their local content in order to get the updated document content
- Whenever a client connects, the backend starts to goroutines (async process): one to read from clients (users) and one to write to clients
   - The write process involes using Go's channels functionality. Essentially, each user has an associated message channel. it is the job of the write channel to pick messages from that user's channel and write a websocket message to that client
   - The read process involves reading messages sent from clients via websockets. Note that clients can only send a message of type operation. So, through a goroutine we continually listen for operaion messages from clients, and handle operations whenever we get them (the manner in which this is done has been described above)
   - These two goroutines also utilize websocket's ping and pong handlers to keep client connections alive, and they are also responsible for cleanup (removing clients from memory) once they disconnect
   - Every two minutes, documents are inserted into the postgres database. The expiry for redis documents is set to 1 hour in order to avoid clogging up memory
   - Go mutexes are used to ensure that shared resources (in memory map of rooms, users etc) are not edited by multiple sources at the time/read during editing
#### Talk to AI
- The Talk to AI feature heavily leaverages the websocket architecture that we developed for the realtime editing
- The challenge with APIs like OpenAI is that they can take some time to generate a response, and we don't want to keep HTTP connections open for long
- This is the flow when a user talks to AI
  - When users click on the talk to AI button, we use the native [SpeechRecognition browser API](https://developer.mozilla.org/en-US/docs/Web/API/SpeechRecognition) to transcribe their speech. Currently, only english (en-US) is supported
  - Then, we send a POST request to `/api/ai` with a body that contains the prompt as well the last known position of the user's cursor, along with the document ID and room code that the user is currently on
  - The server validates the request body, fires off an async goroutine and immediately returns status 202 (Accepted)
  - In the async go routine, we first genrate a prompt for the AI. To make the AI context-aware, we retrive document contents from redis (since we got the docid + room code in the user request), and include 600 characters of content before and after the user's current cursor position, along with the user's original prompt in our OpenAI chat completion api request
  - Once we recieve a response, we send a message of type `operation` with an insert operation consisting of the response we recieved from the OpenAI API to all the users currently connected to that room. The author of this message is the reserved user ID "AI", which the clients utilize to do some frontend work like bringing the talk to ai button back to its idle state from the thinking state

#### Rooms and Documents (creation and retrieval)
-  Users can create rooms that where they can then add documents. Creation of rooms is handled through the API endpoint `POST /api/rooms`. We also have two GET versions of this endpoint to get all public rooms and a specific roomm
-  Once a user joins a room, first we fetch the room info from the backend through the GET api. Once we recieve a room back, we fetch all documents of that room through the GET `/api/documents?roomCode=` endpoint. After getting back the documents, we set the editor to show the first document.
-  Every room will have atleast one document- which is the default document created when the room is created
-  Rooms are identified internally in our database by room codes, which are 6 digit strings.
-  Public rooms are shown on the home page of the app in a paginated list, which private rooms can only be joined if the user knows the room code.
-  Pagination of room data on the home page is achieved through the limit and offset sql queries. We retrive `limit + 1` public rooms from postgres to determine if there are more rooms available, and pass this info down to the frontend so that it can decide if the "show more" button should be disabled

#### Authentication
- The authentication system is pretty standard, we support logging in with username/password and github OAuth. Passwords are hashed, of course.
- On the backend, the authentication system is session based.
### Security
- All API endpoints other than the authentication ones are protected, ie only users with a valid ongoing session can access info related to rooms and documents.
- For websockets, the initial websocket connection is only completed if the user is authentication. In Go, we first get a usual HTTP request which is then upgraded to websocket. The authentication check happens before upgrading to websockets
- We sanitize all user input information that has the potential to be displayed in the UI to prevent cross site scriping attacks
#### Uploading PDFs through GitHub
- Users authenticated with GitHub can upload PDFs to rooms, which are stored in a personal `study-hub-pdfs` repository automatically created on first login
- PDFs are stored at `{roomCode}/{filename}` in the repository, allowing the same filename across different rooms. The GitHub URL is stored in the database for retrieval
- Only the uploader can delete their PDFs; the backend verifies ownership before allowing deletion from both GitHub and the database
- The integration uses GitHub's REST API for repository creation, file upload/retrieval/deletion, with all operations authenticated via the user's personal access token
- PDFs are visible to all room members in the sidebar with owner information and iframe viewer for direct preview

## Deployment

- The app is deployed on the VM with a few updates to the docker+nginx support
- Since the frontend uses dynamic routes (such as website.com/room/{roomCode}/), we had to run the frontend as a node server separately instead of serving static files. This required us to create a Dockerfile from the frontend
- Nginx config needed some updates to allow websockets
- Both the postgres and redis instances are running as docker containers on the VM
- The acme-companion part for HTTPS

## Challenges

1. Realtime communication with websockets and document updates. Architecting the message system was a challenging task, and since this was the first time we were using Go we kind of had to learn both things as we went. Building this system out helped me understand the Go language, operational transform (even though out version is a simplification) and how real-time systems are built
2. Using in-memory databases like Redis in sync with SQL databases. The background sync process was an interesting thing to build, as we knew the high cost of in-memory databases but also the advantage of their speed, which we could see first-hand by how quick the document update operations are.
3. Implementing PDF storage and management through GitHub. We evaluated using AWS S3 versus GitHub for file storage and chose GitHub for better integration with authentication. Configuring OAuth with the `read:user repo` scope required careful handling of permissions and token storage. A key architectural challenge was deciding between storing PDFs per-room or per-user: we ultimately chose per-user repositories to simplify permission management and enable multiple users to upload in a single room. Building the iframe viewer to display PDFs from GitHub's raw content URLs and implementing ownership-based deletion required additional backend verification logic to ensure only uploaders could delete their own files.
