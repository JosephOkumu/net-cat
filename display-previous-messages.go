package main

// Will send every message within the messageHistory slice to a new client.
func displayPreviousMessages(client *Client) {
	historyMu.Lock()
	defer historyMu.Unlock()

	if len(messageHistory) > 0 {
		client.conn.Write([]byte("\n Previous Messages \n"))
		for _, message := range messageHistory {
			client.conn.Write([]byte(message))
		}
		client.conn.Write([]byte(" End Of Previous Messages \n\n"))
	}
}
