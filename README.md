# Critics Finder

Find critics on (Rotten Tomatoes)[https://www.rottentomatoes.com/] that have the same taste as you.

## Build

To build the project, run the following command:
```Bash
go build -o bin/ ./cmd/main.go
```

## Usage

### Fetching the data

First, the data has to be fetched. This data is all the ratings of all the critics.

Use the `fetch` module for fetching the data.

Run the following commands (run subcommands with `-h` flag to see help information):
```Bash
bin/critics_finder fetch critics
bin/critics_finder fetch all-reviews -w 32
```

**Note:** Especially the second command will take some time.

For debugging the subcommand `fetch reviews` is available to fetch the reviews of a specific critic and output some of them to the console.

### Normalizing the data

Because there is no standard rating system on Rotten Tomatoes and everyone does what they want, the ratings need to be normalized. To do this, run
```Bash
bin/critics_finder normalize
```

#### Common rating schemes

To get an idea, how the critics rate the movies, here are some common rating schemes:

- x/10,
- x/5,
- x/4,
- x (since you can't be sure about the scale, I said, that for `x <= 5` it's x/5 and for 5 < x <= 10 it's x/10 and for larger x it's x/100)
- x of 10
- x out of 10
- x stars
- Grades A-F (with and without + and - before or after the letter),
- ...

Needless to say, some ratings are probably normalized incorrectly. But I hope, that the majority of correctly parsed ratings will dominate.

Ratings that couldn't be parsed at all or were I wasn't sure how to handle them, were simply ignored. Examples are

- 2.5.5
- ???
- 1-5 stars
- high +3 out of -4..+4 (whoever wrote this, rated all of their reviews like that)
- Recommended
- ...