# iGoClient
## Implementation memo

### screen draw timing
- User input key to client app
- Client app get server message
    + New message
    + New room
    + User enter
    + User quit

### Concurrency
Use concurrency manager. Get Chanels data and give them to screen.
Use concurrency cancellation signal.

### Problem
How to implement to input Japanese character.
> Resolve
Cursor position is not correct. When inputting multibyte character, 1 unnecessary space can be found.
Furthermore, when using multibyte character and single byte character by turns, many unnecessary space can be gotten.
> Resolve
How to do when we need to have many times conversation? Receive function should judge each flow?

### TODO
- implement tcp connection 
    - Ping function
    - Receive function
    - Send function
	- 

chrome 検索機能 staff search jira URI query で検索
