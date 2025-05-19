# SVG Web Elements

A project for getting SVG illustrations from a URL to use within documentation.

## Features

- Attributes to manipulate the contents and looks of the SVG
- Ability to insert via markdown image or HTML `<img>` tag
- Customizable text, colors, and dimensions

## How it works

This project exposes a web service where you can get an SVG illustration from a URL like:
```
https://this-service/ui/basic-auth.svg?url=https://example.com
```

## Usage

### Running the server

```bash
# Navigate to the project directory
cd svg-web-elements

# Run the server
go run cmd/server/main.go
```

The server will start on port 8082 by default. You can customize the port with the `PORT` environment variable.

### Accessing SVGs

Access SVGs through the `/ui/` endpoint:

```
http://localhost:8082/ui/basic-auth.svg
```

### Customization Parameters

You can customize the SVGs using query parameters:

- `width` - Set the SVG width (e.g., `width=400`). When specified alone, height scales proportionally.
- `height` - Set the SVG height (e.g., `height=200`). When specified alone, width scales proportionally.
- `text.{element-id}` - Replace text in element with ID (e.g., `text.text-title=Login`)
- `color.{element-id}` - Change color of element with ID (e.g., `color.page-background=%23f0f9ff`) - Note: Use `%23` instead of `#` in URLs for hex colors
- `url` - External URL to display (e.g., `url=https://example.com`)

### Examples

Basic usage:
```
http://localhost:8082/ui/basic-auth.svg
```

Customized text:
```
http://localhost:8082/ui/basic-auth.svg?text.text-title=Login&text.text-url=example.com
```

Customized size (proportional scaling):
```
http://localhost:8082/ui/basic-auth.svg?width=400&height=200
```

Single dimension (automatically maintains aspect ratio):
```
http://localhost:8082/ui/basic-auth.svg?width=400
```

Customized colors:
```
http://localhost:8082/ui/basic-auth.svg?color.page-background=%23f0f9ff&color.btn-background_2=%230ea5e9
```

## Project Structure

```
svg-web-elements/
├── cmd/
│   └── server/
│       └── main.go           # Entry point for the server
├── internal/
│   ├── handlers/             # HTTP handlers
│   └── svg/                  # SVG processing logic
└── static/
    └── svg/                  # SVG templates
        └── basic-auth.svg    # Example SVG
```

## Advanced Features

- **Proportional Scaling**: Specify either width or height, and the other dimension will scale automatically to maintain the aspect ratio
- **SVG Diagnostics**: Access `/debug?svg=basic-auth.svg` to inspect SVG elements and their IDs
- **Element Customization**: Modify text and colors by targeting specific element IDs

## Languages and Technologies

Project is written in Go with plain HTML, CSS, and JavaScript for the web interface.

## Docker Deployment

This project can be deployed using Docker and Docker Compose.

### Building and Running with Docker Compose

1. Clone the repository
```bash
git clone https://github.com/yourusername/svg-web-elements.git
cd svg-web-elements
```

2. Create a `.env` file with your environment variables:
```bash
DOMAIN=example.com  # Your domain
PUID=1000           # User ID
PGID=1000           # Group ID
TZ=UTC              # Timezone
```

3. Build and start the container:
```bash
docker compose up -d
```

The service will be available at `https://svg.example.com` (or whatever domain you configured).

### Manual Docker Build

If you prefer to build and run the Docker image manually:

```bash
# Build the image
docker build -t svg-web-elements .

# Run the container
docker run -d -p 8082:8082 -v $(pwd)/static:/app/static -v $(pwd)/svg-cache:/app/cache --name svg-web-elements svg-web-elements
```

The service will be available at `http://localhost:8082`.

### Configuration

When running in Docker, you can configure the application using environment variables:
- `PORT`: The port the application listens on (default: 8082)
- `HOST`: The host interface to bind to (default: "" which binds to all interfaces)
- `SVG_DIR`: The base directory for the application (default: auto-detected)
- `TZ`: Timezone
- `PUID`/`PGID`: User and group IDs for file permissions

For convenience, you can also use the included start script:
```bash
./start.sh
```
