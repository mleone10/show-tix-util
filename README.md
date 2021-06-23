# ShowTix4U Utility

This program uses the ShowTix4U Transaction Search API to download transactions for a given event. It then parses that event's transactions into a format that can be interpretted by SaasAnt's import tool.

## Usage

To use the program, first log in to ShowTix4U and use Chrome's developer tools to extract the target event's ID and the authorization token contained in the `connect.sid` cookie. Then, run the program like so:

```bash
go run main.go --event $EVENT_ID --token $TOKEN
```

The program will print CSV-formatted line items to STDOUT.

## To Do

- ShowTix4U currently charges a $1.50 fee for comp tickets. This means that transactions containing comp tickets may end up with negative net totals, which Quickbooks does not allow. Those transactions must be instead registered as expenses and imported separately.
