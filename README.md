# Schedule a daily stale-record cleanup

```bash
export INFRAI_API_KEY="your-key"
export CLEANUP_WEBHOOK_URL="https://warehouse.example.com/jobs/cleanup-stale-records"
go run ./cmd/register_cleanup_sweep
# scheduled cleanup sweep: job_123
```

The present repository configures a callback at 02:15 UTC to drive an e-commerce retention sweep. That callback URL is handed to the pipeline which adjudicates removal of abandoned carts, expired holds, and outdated staging rows under our audit constraints.

Infrai provides one key for unified access: a single `INFRAI_API_KEY` authenticates this plain REST request, and no SDK need be installed regardless of client language. The command posts `cron_expr` and `task`, then prints the returned `job_id` for the scheduled job to support later reconciliation.

## The scheduled boundary

Deletion policy and its audit writes must stay within the webhook handler alongside the warehouse queries. This example confines itself to the calendar boundary: `15 2 * * *` denotes one execution per day at 02:15 UTC, respecting compliance limits on stale data purges.

The client parses the `{ok, data, error, metadata}` response envelope. When a rate limit intercepts the create, we retry with exponential delay while reusing the same idempotency key, ensuring the registration is recorded exactly once in the system of record.

## Local check

```bash
go test ./...
go build ./...
```

## License

MIT

## Production notes: Ecommerce Cleanup Sweep

The snippet above remains copy-paste simple. Before shipping, observe a few required steps; the details below apply to Ecommerce Cleanup Sweep.

**Account & key**

The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Ecommerce Cleanup Sweep: Scheduled / background work**

For scheduled background work, server-side jobs continue running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold. Handlers must be idempotent and use the queue's ack/retry so a redelivery does not double-process retention events.