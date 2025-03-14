# Setting Up Real-Time Synchronization with Supabase Webhooks

This guide explains how to configure Supabase to send real-time updates to your blockchain ledger service whenever changes occur in your database.

## Overview

Instead of relying solely on periodic synchronization, the blockchain ledger service now supports real-time updates through Supabase webhooks. When records are inserted, updated, or deleted in your Supabase database, a webhook notification is sent to your service, which immediately processes the change and creates blockchain transactions as needed.

## Prerequisites

- A running instance of the blockchain ledger service
- Admin access to your Supabase project
- (Optional) A secure environment for webhook signature verification

## Configuration Steps

### 1. Set Up Environment Variables

Add the following to your `.env` file:

```
WEBHOOK_SECRET=your_secure_random_string
```

This secret will be used to verify that webhook requests are coming from your Supabase instance.

### 2. Configure Supabase Database Webhooks

1. Log in to your Supabase dashboard
2. Navigate to Database → Webhooks
3. Click "Create a new webhook"
4. Configure the webhook with the following settings:
   - **Name**: `blockchain_ledger_sync`
   - **Table**: Select both `drugs` and `shipments` tables
   - **Events**: Select `INSERT`, `UPDATE`, and optionally `DELETE`
   - **URL**: Enter the URL of your blockchain ledger service's webhook endpoint: `https://your-service-url.com/api/webhooks/supabase`
   - **HTTP Method**: `POST`
   - **Headers**: Add `X-Webhook-Signature` with the value matching your `WEBHOOK_SECRET`

5. Click "Save"

### 3. Test the Webhook

1. Make a change to a record in either the `drugs` or `shipments` table in Supabase
2. Check your blockchain ledger service logs to confirm that the webhook was received and processed
3. Verify that a new blockchain transaction was created for the changed record

## How It Works

When a record is changed in Supabase:

1. Supabase sends a webhook notification to your service's `/api/webhooks/supabase` endpoint
2. The webhook handler verifies the request signature
3. The handler extracts the changed record data
4. A blockchain transaction is created for the record (if needed)
5. The record is saved to the local data store with its blockchain transaction ID
6. The sync status is updated to reflect the real-time change

## Troubleshooting

### Webhook Not Triggering

- Verify that your service is publicly accessible at the URL configured in Supabase
- Check that the webhook is enabled in Supabase
- Ensure the tables and events are correctly configured

### Authentication Errors

- Confirm that the `WEBHOOK_SECRET` in your `.env` file matches the `X-Webhook-Signature` header value in Supabase
- Check your service logs for signature verification errors

### Processing Errors

- Look for errors in your service logs related to record processing or blockchain transaction creation
- Verify that the webhook payload structure matches what your handler expects

## Benefits of Real-Time Synchronization

- Immediate blockchain transaction creation when data changes
- Reduced latency between database updates and blockchain ledger updates
- More efficient resource usage compared to periodic polling
- Better tracking of the exact time when changes occurred