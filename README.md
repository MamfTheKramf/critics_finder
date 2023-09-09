# Critics Finder

Find critics on (Rotten Tomatoes)[https://www.rottentomatoes.com/] that have the same taste as you.

## Usage

### Fetching the data

First, the data has to be fetched. This data is all the ratings of all the critics.

Use the `fetch` module for fetching the data.

Run the following commands (run subcommands with `-h` flag to see help information):
```
go run . fetch critics -o ./tmp/critivs.csv
go run . fetch all-reviews -i ./tmp/critics.csv -o ./tmp/reviews -w 32
```

**Note:** Especially the second command will take some time.

For debugging the subcommand `fetch reviews` is available to fetch the reviews of a specific critic and output some of them to the console.