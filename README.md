# Net-Cat
This Golang program is a simple implementation of a chat system with a server-client architecture. It allows multiple clients to connect to a server over a TCP connection, send messages to each other in real-time, and view messages from all other connected clients. This project mimics the functionality of the well-known NetCat utility but with a focus on chat functionality and user interaction.
## Documentation

This section illustrates how to make use of this program.

### Installation

To run this program, download and install the latest version of Go from [here](https://go.dev/doc/install).

### Usage

1. Clone this repository onto your terminal by using the following command:
    ```bash
    git clone https://learn.zone01kisumu.ke/git/hilaokello/net-cat
    ```

2. Navigate into net-cat directory by using the command:
    ```bash
    cd net-cat
    ```

3. To start the server, execute the command below:
    ```bash
    go run . 
    ```
   This will by default start the server on port 8989
   You can specify a port number by adding the port number as shown below:

   Example:
   ```bash
   go run . 8000
   ```
   To connect a client to the server use the nc command below followed by the port number:
   ```bash
   nc localhost 8989
   ```
4. In the prompt displayed, enter your username.
5. Enter your message to start group chatting with other users.
### Features
- Server-client architecture.
- Real-time group chat.
- Multiple users can join and leave the chat. Supports a maximum    of 10 connected clients.
- Efficient concurrency with go routines.

### Contributions

Pull requests are welcome! Users of this program are encouraged to contribute by adding features or fixing bugs. For major changes, please open an issue first to discuss your ideas.

### Authors
[josotieno](https://learn.zone01kisumu.ke/git/josotieno/)

[hilaokello](https://learn.zone01kisumu.ke/git/hilaokello)

[johnoodhiamboO](https://learn.zone01kisumu.ke/git/johnodhiambo0)

## Licence
[MIT License](./LICENSE)
## Credits
[Zone01Kisumu](https://www.zone01kisumu.ke/)

