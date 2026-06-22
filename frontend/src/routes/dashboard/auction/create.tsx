import { useCreateAuction } from '#/hooks/use-auction'
import { createFileRoute } from '@tanstack/react-router'
import { useForm } from '@tanstack/react-form'
import { AuctionType } from '#/gen/auction_platform/v1/auction_pb'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from '#/components/ui/card'
import { Input } from '#/components/ui/input'
import { Button } from '#/components/ui/button'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Switch } from '#/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select'
import { Alert, AlertDescription } from '#/components/ui/alert'
import { Badge } from '#/components/ui/badge'

export const Route = createFileRoute('/dashboard/auction/create')({
  component: RouteComponent,
})

function RouteComponent() {
  const {
    mutate: createAuction,
    isPending,
    error: createAuctionError,
  } = useCreateAuction()

  const form = useForm({
    defaultValues: {
      title: '',
      description: '',
      type: AuctionType.ENGLISH as AuctionType,
      starting_price: 0,
      reserved_price: 0,
      extend_on_bid: false,
      extend_minutes: 5,
      start_time: '',
      end_time: '',
      drop_amount: 0,
      drop_interval: 0,
    },
    onSubmit: async ({ value }) => {
      createAuction({ data: value })
    },
  })

  return (
    <div className="min-h-screen py-10 px-4 bg-muted/30">
      <div className="max-w-3xl mx-auto">
        {/* Page header */}
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              List an auction
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              Buyers will see everything you enter below.
            </p>
          </div>
          <Badge variant="secondary" className="shrink-0 mt-1">
            Prices in cents
          </Badge>
        </div>

        {createAuctionError && (
          <Alert variant="destructive" className="mb-6">
            <AlertDescription>
              {createAuctionError instanceof Error
                ? createAuctionError.message
                : 'Failed to create auction.'}
            </AlertDescription>
          </Alert>
        )}

        <form
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            form.handleSubmit()
          }}
          className="space-y-5"
        >
          {/* ── Card 1 — Item details ───────────────────────────── */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Item details</CardTitle>
              <CardDescription>
                Title and description shown to bidders
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <form.Field
                name="title"
                validators={{
                  onChange: ({ value }) =>
                    !value.trim()
                      ? 'Title is required'
                      : value.length < 3
                        ? 'At least 3 characters'
                        : undefined,
                }}
              >
                {(field) => (
                  <div className="space-y-1.5">
                    <Label htmlFor={field.name}>Title</Label>
                    <Input
                      id={field.name}
                      name={field.name}
                      type="text"
                      value={field.state.value}
                      placeholder="e.g. Vintage 1965 Rolex Submariner"
                      aria-invalid={field.state.meta.errors.length > 0}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                    />
                    {field.state.meta.errors.length > 0 && (
                      <p className="text-xs text-destructive">
                        {String(field.state.meta.errors[0])}
                      </p>
                    )}
                  </div>
                )}
              </form.Field>

              <form.Field
                name="description"
                validators={{
                  onChange: ({ value }) =>
                    !value.trim()
                      ? 'Description is required'
                      : value.length < 5
                        ? 'At least 5 characters'
                        : undefined,
                }}
              >
                {(field) => (
                  <div className="space-y-1.5">
                    <Label htmlFor={field.name}>Description</Label>
                    <Textarea
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      rows={3}
                      placeholder="Condition, provenance, what's included…"
                      aria-invalid={field.state.meta.errors.length > 0}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                    />
                    {field.state.meta.errors.length > 0 && (
                      <p className="text-xs text-destructive">
                        {String(field.state.meta.errors[0])}
                      </p>
                    )}
                  </div>
                )}
              </form.Field>
            </CardContent>
          </Card>

          {/* ── Card 2 — Format & pricing ───────────────────────── */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Format & pricing</CardTitle>
              <CardDescription>
                English auctions go up, Dutch auctions go down
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <form.Field name="type">
                {(field) => (
                  <div className="space-y-1.5">
                    <Label htmlFor="auction-type">Auction type</Label>
                    <Select
                      value={String(field.state.value)}
                      onValueChange={(val) => field.handleChange(Number(val))}
                    >
                      <SelectTrigger id="auction-type" className="w-full">
                        <SelectValue placeholder="Select auction type" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value={String(AuctionType.ENGLISH)}>
                          English — price goes up, highest bid wins
                        </SelectItem>
                        <SelectItem value={String(AuctionType.DUTCH)}>
                          Dutch — price drops, first to accept wins
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
              </form.Field>

              <div className="grid grid-cols-2 gap-4">
                <form.Field
                  name="starting_price"
                  validators={{
                    onChange: ({ value }) =>
                      !value || Number(value) <= 0
                        ? 'Must be greater than 0'
                        : undefined,
                  }}
                >
                  {(field) => (
                    <div className="space-y-1.5">
                      <Label htmlFor={field.name}>Starting price (¢)</Label>
                      <Input
                        id={field.name}
                        name={field.name}
                        type="number"
                        min={1}
                        value={field.state.value}
                        placeholder="500"
                        aria-invalid={field.state.meta.errors.length > 0}
                        onChange={(e) =>
                          field.handleChange(Number(e.target.value))
                        }
                        onBlur={field.handleBlur}
                      />
                      {field.state.meta.errors.length > 0 && (
                        <p className="text-xs text-destructive">
                          {String(field.state.meta.errors[0])}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>

                <form.Field
                  name="reserved_price"
                  validators={{
                    onChange: ({ value }) =>
                      Number(value) < 0 ? 'Must be 0 or greater' : undefined,
                  }}
                >
                  {(field) => (
                    <div className="space-y-1.5">
                      <Label htmlFor={field.name}>Reserve price (¢)</Label>
                      <Input
                        id={field.name}
                        name={field.name}
                        type="number"
                        min={0}
                        value={field.state.value}
                        placeholder="0 = no reserve"
                        aria-invalid={field.state.meta.errors.length > 0}
                        onChange={(e) =>
                          field.handleChange(Number(e.target.value))
                        }
                        onBlur={field.handleBlur}
                      />
                      {field.state.meta.errors.length > 0 && (
                        <p className="text-xs text-destructive">
                          {String(field.state.meta.errors[0])}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>
              </div>

              {/* Dutch-only fields — appear inline, same card */}
              <form.Subscribe selector={(s) => s.values.type}>
                {(type) =>
                  type === AuctionType.DUTCH ? (
                    <div className="grid grid-cols-2 gap-4 rounded-lg border bg-muted/40 p-4">
                      <form.Field
                        name="drop_amount"
                        validators={{
                          onChange: ({ value }) =>
                            value <= 0 ? 'Must be positive' : undefined,
                        }}
                      >
                        {(field) => (
                          <div className="space-y-1.5">
                            <Label htmlFor={field.name}>Drop amount (¢)</Label>
                            <Input
                              id={field.name}
                              name={field.name}
                              type="number"
                              min={1}
                              value={field.state.value}
                              placeholder="100"
                              aria-invalid={field.state.meta.errors.length > 0}
                              onChange={(e) =>
                                field.handleChange(Number(e.target.value))
                              }
                              onBlur={field.handleBlur}
                            />
                            {field.state.meta.errors.length > 0 && (
                              <p className="text-xs text-destructive">
                                {String(field.state.meta.errors[0])}
                              </p>
                            )}
                          </div>
                        )}
                      </form.Field>

                      <form.Field
                        name="drop_interval"
                        validators={{
                          onChange: ({ value }) =>
                            value <= 0 ? 'Must be positive' : undefined,
                        }}
                      >
                        {(field) => (
                          <div className="space-y-1.5">
                            <Label htmlFor={field.name}>
                              Drop every (seconds)
                            </Label>
                            <Input
                              id={field.name}
                              name={field.name}
                              type="number"
                              min={1}
                              value={field.state.value}
                              placeholder="30"
                              aria-invalid={field.state.meta.errors.length > 0}
                              onChange={(e) =>
                                field.handleChange(Number(e.target.value))
                              }
                              onBlur={field.handleBlur}
                            />
                            {field.state.meta.errors.length > 0 && (
                              <p className="text-xs text-destructive">
                                {String(field.state.meta.errors[0])}
                              </p>
                            )}
                          </div>
                        )}
                      </form.Field>
                    </div>
                  ) : null
                }
              </form.Subscribe>
            </CardContent>
          </Card>

          {/* ── Card 3 — Schedule & anti-snipe ──────────────────── */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Schedule</CardTitle>
              <CardDescription>When bidding opens and closes</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <form.Field
                  name="start_time"
                  validators={{
                    onChange: ({ value }) =>
                      !value ? 'Start time is required' : undefined,
                  }}
                >
                  {(field) => (
                    <div className="space-y-1.5">
                      <Label htmlFor={field.name}>Starts at</Label>
                      <Input
                        id={field.name}
                        name={field.name}
                        type="datetime-local"
                        value={field.state.value}
                        aria-invalid={field.state.meta.errors.length > 0}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                      />
                      {field.state.meta.errors.length > 0 && (
                        <p className="text-xs text-destructive">
                          {String(field.state.meta.errors[0])}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>

                <form.Field
                  name="end_time"
                  validators={{
                    onChange: ({ value, fieldApi }) => {
                      if (!value) return 'End time is required'
                      const start = fieldApi.form.getFieldValue('start_time')
                      if (start && new Date(value) <= new Date(start)) {
                        return 'Must be after start time'
                      }
                      return undefined
                    },
                  }}
                >
                  {(field) => (
                    <div className="space-y-1.5">
                      <Label htmlFor={field.name}>Ends at</Label>
                      <Input
                        id={field.name}
                        name={field.name}
                        type="datetime-local"
                        value={field.state.value}
                        aria-invalid={field.state.meta.errors.length > 0}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                      />
                      {field.state.meta.errors.length > 0 && (
                        <p className="text-xs text-destructive">
                          {String(field.state.meta.errors[0])}
                        </p>
                      )}
                    </div>
                  )}
                </form.Field>
              </div>

              <form.Field name="extend_on_bid">
                {(field) => (
                  <div className="flex items-center justify-between rounded-lg border p-3.5">
                    <div className="space-y-0.5 pr-4">
                      <Label
                        htmlFor={field.name}
                        className="text-sm font-medium cursor-pointer"
                      >
                        Extend on late bid
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Adds time when a bid lands near the deadline
                      </p>
                    </div>
                    <Switch
                      id={field.name}
                      checked={field.state.value}
                      onCheckedChange={(checked) => field.handleChange(checked)}
                    />
                  </div>
                )}
              </form.Field>

              <form.Subscribe selector={(s) => s.values.extend_on_bid}>
                {(extendOnBid) =>
                  extendOnBid ? (
                    <form.Field
                      name="extend_minutes"
                      validators={{
                        onChange: ({ value }) =>
                          value <= 0 ? 'Must be at least 1 minute' : undefined,
                      }}
                    >
                      {(field) => (
                        <div className="space-y-1.5 max-w-[200px]">
                          <Label htmlFor={field.name}>
                            Extend by (minutes)
                          </Label>
                          <Input
                            id={field.name}
                            name={field.name}
                            type="number"
                            min={1}
                            max={60}
                            value={field.state.value}
                            aria-invalid={field.state.meta.errors.length > 0}
                            onChange={(e) =>
                              field.handleChange(Number(e.target.value))
                            }
                            onBlur={field.handleBlur}
                          />
                          {field.state.meta.errors.length > 0 && (
                            <p className="text-xs text-destructive">
                              {String(field.state.meta.errors[0])}
                            </p>
                          )}
                        </div>
                      )}
                    </form.Field>
                  ) : null
                }
              </form.Subscribe>
            </CardContent>

            <CardFooter>
              <form.Subscribe
                selector={(s) => ({
                  canSubmit: s.canSubmit,
                  isSubmitting: s.isSubmitting,
                })}
              >
                {({ canSubmit, isSubmitting }) => (
                  <Button
                    type="submit"
                    className="w-full"
                    disabled={!canSubmit || isSubmitting || isPending}
                  >
                    {isSubmitting || isPending
                      ? 'Creating auction…'
                      : 'Create auction'}
                  </Button>
                )}
              </form.Subscribe>
            </CardFooter>
          </Card>
        </form>
      </div>
    </div>
  )
}
