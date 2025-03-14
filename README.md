# Blockchain Ledger Service - Go Implementation

This is a Go implementation of the Blockchain Ledger Service for tracking drugs and shipments. It provides improved performance over the Python version while maintaining the same functionality.

## Setup

1. Create a `.env` file in the `go_version` directory with your Supabase credentials:

   ```
   SUPABASE_URL=your_supabase_url
   SUPABASE_KEY=your_supabase_anon_key
   SUPABASE_SERVICE_KEY=your_supabase_service_key
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

## Running the Service

1. Start the main application:

   ```bash
   go run main.go
   ```

   The server will start on port 3000 by default. You can change this by setting the `PORT` environment variable.
   
   The service will automatically synchronize data with Supabase every 5 minutes by default. You can change this interval by setting the `SYNC_INTERVAL` environment variable (e.g., `SYNC_INTERVAL=10m` for 10 minutes).

## API Endpoints

### Drug Endpoints

- `POST /api/drugs` - Create a new drug record
- `GET /api/drugs` - Get all drug records
- `GET /api/drugs/:id` - Get a specific drug record
- `PUT /api/drugs/:id` - Update a drug record

### Shipment Endpoints

- `POST /api/shipments` - Create a new shipment record
- `GET /api/shipments` - Get all shipment records
- `GET /api/shipments/:id` - Get a specific shipment record
- `PUT /api/shipments/:id` - Update a shipment record

### Blockchain Endpoints

- `GET /api/blockchain/status` - Get the current status of the blockchain
- `GET /api/blockchain/verify/:tx_hash` - Verify a blockchain transaction
- `POST /api/blockchain/consistency-check` - Run a consistency check on the blockchain

### Synchronization Endpoints

- `GET /api/sync/status` - Get the current status of the synchronization service
- `POST /api/sync/force` - Force an immediate synchronization with Supabase

## Service Key Importance

The Supabase service key is essential for this application to function correctly. Here's why:

- The blockchain ledger needs to bypass Row Level Security (RLS) policies to update drug and shipment records
- The service key allows the application to perform database operations without authentication restrictions
- Without the service key, operations like updating drug records or processing shipments may fail

You can find your service key in the Supabase dashboard:

1. Go to Project Settings
2. Click on API
3. Look for "service_role key" or "service key"

**Important**: Keep your service key secure and never expose it in your code or version control.

## Security

- Uses environment variables for sensitive credentials
- Implements service key for secure database operations
- Validates input data before processing
- Handles errors gracefully without exposing sensitive information

## Performance Improvements

The Go implementation offers several performance improvements over the Python version:

- Concurrent request handling with Go's lightweight goroutines
- Efficient memory management
- Faster JSON serialization/deserialization
- Reduced latency for blockchain operations
- Improved database connection handling
- Automatic synchronization with Supabase for drugs and shipments data

## Building for Production

To build the application for production:

```bash
go build -o blockchain_ledger_service
```

This will create an executable file that you can run directly:

```bash
./blockchain_ledger_service
```