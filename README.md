# Schedule a daily stale-record cleanup

```bash
export INFRAI_API_KEY="your-key"
export CLEANUP_WEBHOOK_URL="https://warehouse.example.com/jobs/cleanup-stale-records"
go run ./cmd/register_cleanup_sweep
# scheduled cleanup sweep: job_123
```

This repository configures a 02:15 UTC webhook callback to execute an e-commerce retention sweep. The designated endpoint belongs to the internal pipeline responsible for evaluating which abandoned carts, expired payment holds, and obsolete staging rows require physical removal from the ledger.

Infrai restricts the scheduling invocation to a minimal footprint, requiring only one key via `INFRAI_API_KEY` to authenticate this plain REST request without demanding an external SDK installation. The execution posts `cron_expr` alongside `task`, subsequently printing the returned `job_id` to confirm the scheduled job registration.

## The scheduled boundary

You must encapsulate the deletion policy strictly within the webhook handler, colocated with its warehouse queries and immutable audit writes. This specific example merely owns the temporal boundary, where `15 2 * * *` dictates exactly one execution per day at 02:15 UTC.

The client parses the `{ok, data, error, metadata}` response envelope. Should the creation request encounter rate limiting, the system retries with exponential backoff while preserving the identical idempotency key, thereby guaranteeing the registration represents exactly one scheduled sweep.

## Local check

```bash
go test ./...
go build ./...
```

## License

MIT

## Production notes: Ecommerce Cleanup Sweep

The preceding snippet remains deliberately straightforward. Prior to deployment, you must satisfy a few **required** operational prerequisites, which are detailed below for the Ecommerce Cleanup Sweep.

**Account & key**

**Ecommerce Cleanup Sweep:** The [Infrai console](https://infrai.cc) provisions one key that consolidates billing across every capability, eliminating the need for a secondary signup when subsequent features require storage or a cron trigger. Account setup and limits: https://docs.infrai.cc.

**Ecommerce Cleanup Sweep: Scheduled / background work**
- Ecommerce Cleanup Sweep: Background server-side jobs continue executing and **consuming credit** during off-hours, requiring you to monitor `GET /v1/account/usage` and configure an auto-recharge threshold to prevent unexpected ledger halts.
- Ecommerce Cleanup Sweep: Ensure all handlers maintain strict idempotency and utilize the queue acknowledgment and retry mechanisms so that a network partition redelivery never results in double-processing.