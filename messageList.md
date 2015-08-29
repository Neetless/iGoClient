# Socket communication
## Send message list
### LOGIN
- LOGIN %81%DC%2A%28%81E%81%CD%81E%29%2A%81%DC
LOGIN ⌒*(・∀・)*⌒
LOGIN testUser
- SET_INTRO 高町なのは小学３年生９歳です
SET_INTRO testIntro
- SET_LEVEL 不明
SET_LEVEL testLevel
- SET_ID 58
- CLIENT_INFO lilical8.0_magical_nanoha

### PING
- OK SVR_PING
Client recieve the message "SVR_PING" from server, then send "OK SVR_PING"
- PING -1
Duration 360000ms

### ROOM
- SEND: ADD_ROOM 5|6|9|13|19 GAME|CHAT 任意の部屋名|検討用|対局用[5|6|9|13|19]
ADD_ROOM 19 GAME 任意の部屋名[19]
- SEND: GET_ROOM_INFO 1|2|....
Room # is s start from 1
Response: 
Server response: USERS 1 test:
Server response: ROOM_INFO 1 test null null GAME_PREPARE 6.5 test[19]
- Response: ROOM_INFO 1 test2 null null GAME_PREPARE 6.5 test[19]
test2 is owner of the room.
- Response: ROOM_REMOVED 1|2|....
- Response: LEAVE <ID> <USER> (like: LEAVE 0 test)
- Send: OPEN_ROOM <room #>
- Send: CLOSE_ROOM <room #>
- RESPONCE: OK CLOSE_ROOM <room #>
- Responce: ENTER <id> <user>
- Responce ROOM_ADDED <id> <owner> <room name> null?

# Chat
- SEND: "SHOUT " + this.room_id + " " + localButton.getLabel().replace(' ', '\t')
- SEND: send("SHOUT " + paramInt + " " + paramString);
- RESPONSE: MESSAGE <room #> <text>

# User
- Responce: ENTER <room #> <user>(like: ENTER 0 test)
- Responce: ENTER 0 <user>
When user login, this response come.
- Response LEAVE <room #> <user>
- LEAVE 0 <user>
Other user logout.

