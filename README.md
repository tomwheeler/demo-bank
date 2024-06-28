# demo-bank

This project contains code to run banking services (and a corresponding
GUI that will show account balances and the service status). It is 
intended for use in demonstrations, particularly the money transfer 
example. It is written in Go and uses no Temporal APIs. Although it 
provides a `BankClient` class that can be used in a Temporal Workflow, 
presumably you could implement something similar in another language 
if you need to demonstrate something that involves account balance 
lookups, deposits, and withdrawals.

There are three commands:

* One for starting the banking service for the person sending money
  (default: Tom, runs on port 8888).
* One for starting the banking service for the person receiving money 
  (default: Ted, runs on port 8889).
* One for launching the GUI

Examples of how to run those commands is shown below. It is possible 
to change the names and port numbers through command-line options.

# Start the Service for the Sender's Bank

```bash
go run ./cmd/sender-banking-service/
```

# Start the Service for the Recipient's Bank

```bash
go run ./cmd/recipient-banking-service/
```

# Launch the Banking UI

```bash
go run ./cmd/ui
```
Due to a [known issue](https://github.com/fyne-io/fyne/issues/4502) 
with the Fyne package, you may see a `ld: warning: ignoring duplicate 
libraries: '-lobjc'` warning when running this command. You can ignore 
it, but it's possible to use the following command to avoid it:

```bash
go run -ldflags="-extldflags=-Wl,-ld_classic" ./cmd/ui
```

# Manually Depositing or Withdrawing Money

Once a given bank service has been started, you can use an HTTP
client to add or remove money from that account. The examples 
below assume that you're running the bank service locally and 
on the default ports.

```bash
# Deposit $1100 into the sender's account 
curl http://localhost:8888/deposit?amount=1100
curl http://localhost:8888/withdraw?amount=100
  
# Deposit $5000 into the recipient's account 
curl http://localhost:8889/deposit?amount=5000

# Deposit $5000 into the recipient's account, but use an idempotency
# key to prevent a duplicate request from being processed
curl http://localhost:8889/deposit?amount=5000&idempotency-key=12345
```

Alternatively, you can modify the balance in the data file 
associated with that bank, although you'll then need to 
restart the service for the change to take effect.
