# TMDB CLI

A command-line interface (CLI) for The Movie Database (TMDB), inspired by the [roadmap.sh tmdb-cli project idea](https://roadmap.sh/projects/tmdb-cli).

> Project idea from: https://roadmap.sh/projects/tmdb-cli

## Project Idea
This project is a CLI tool that allows users to fetch and display movie data from TMDB directly from their terminal. It supports different movie categories such as now playing, popular, top rated, and upcoming movies. The CLI is built using Go and the Cobra library for easy command and flag management.

## Features
- Fetch and display lists of movies from TMDB
- Supported categories:
  - Now Playing
  - Popular
  - Top Rated
  - Upcoming
- Pretty-printed output of movie titles
- API authentication using a Bearer token (TMDB API key)
- Simple flag-based interface

## Usage

### Prerequisites
- Go installed (version 1.18+ recommended)
- TMDB API key (get one from [TMDB API](https://developers.themoviedb.org/3/getting-started/introduction))

### Setup
1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd tmdb
   ```
2. Set your TMDB API key as an environment variable:
   ```bash
   export TMDB_API_KEY=your_api_key_here
   ```
3. Run the CLI:
   ```bash
   go run main.go --type playing
   go run main.go --type popular
   go run main.go --type top_rated
   go run main.go --type upcoming
   ```

### Flags
- `--type` or `-t`: Specify the movie category to fetch. Valid values are `playing`, `popular`, `top_rated`, `upcoming`.

### Example Output
```
Now Playing Movies:
1. War of the Worlds
2. Jurassic World Rebirth
3. William Tell
...
```

## Project Structure
```
tmdb/
├── main.go           # Entry point
├── cmd/              # Cobra command definitions
│   └── root.go
├── tmdbapi/          # API logic for TMDB
│   └── api.go
├── go.mod
├── go.sum
└── README.md
```

## How It Works
- The CLI parses the `--type` flag and calls the corresponding function in `tmdbapi/api.go`.
- The API logic makes an authenticated HTTP request to TMDB, parses the JSON response, and prints a pretty list of movie titles.
- All API logic is kept separate from command logic for maintainability.

## Contributing
Pull requests and suggestions are welcome! Please open an issue or PR for improvements.

## License
MIT

---
Inspired by [roadmap.sh tmdb-cli project idea](https://roadmap.sh/projects/tmdb-cli)
