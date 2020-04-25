extern crate futures;
extern crate websocket;
extern crate tokio;

use websocket::client::ClientBuilder;
use websocket::{Message, OwnedMessage};
use std::fmt::Display;
use std::string::toString;

const CONNECTION: &'static str = "wss://md-api.tradovate.com/v1/websocket?r=0.8840574374908023";

fn main() {
	println!("Connecting to {}", CONNECTION);

	let mut client = ClientBuilder::new(CONNECTION)
		.unwrap()
		.add_protocol("rust-websocket")
		.connect(None)
		.unwrap();

    println!("Successfully connected");

    let message = Message::text("authorize\n2\n\n");
    client.send_message(&message).unwrap();    

    let response = client.recv_message().unwrap();
}