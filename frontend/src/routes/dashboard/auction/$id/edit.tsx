import { createFileRoute, Link } from '@tanstack/react-router'
import { DashboardLayout } from '#/components/dashboard-layout'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  CardFooter,
} from '#/components/ui/card'
import { Button } from '#/components/ui/button'
import { Input } from '#/components/ui/input'
import { Label } from '#/components/ui/label'
import { Textarea } from '#/components/ui/textarea'
import { Badge } from '#/components/ui/badge'
import { ArrowLeft, Save } from 'lucide-react'

export const Route = createFileRoute('/dashboard/auction/$id/edit')({
  component: EditAuctionPage,
})

function EditAuctionPage() {
  const { id } = Route.useParams()

  return (
    <DashboardLayout>
      <div className="max-w-3xl mx-auto space-y-6">
        <div>
          <Link
            to="/dashboard/auction/$id"
            params={{ id }}
            className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors mb-3"
          >
            <ArrowLeft className="size-3.5" />
            Back to auction
          </Link>
          <h1 className="text-2xl font-bold tracking-tight">Edit auction</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Update title, description, schedule or pricing
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Item details</CardTitle>
            <CardDescription>Modify the listing information</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1.5">
              <Label>Title</Label>
              <Input placeholder="Auction title" />
            </div>
            <div className="space-y-1.5">
              <Label>Description</Label>
              <Textarea rows={4} placeholder="Describe your item" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>Starting price (¢)</Label>
                <Input type="number" placeholder="500" />
              </div>
              <div className="space-y-1.5">
                <Label>Reserve price (¢)</Label>
                <Input type="number" placeholder="0 = no reserve" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>Starts at</Label>
                <Input type="datetime-local" />
              </div>
              <div className="space-y-1.5">
                <Label>Ends at</Label>
                <Input type="datetime-local" />
              </div>
            </div>
          </CardContent>
          <CardFooter className="flex justify-between">
            <Badge variant="secondary" className="text-xs">
              Auction ID: {id}
            </Badge>
            <Button>
              <Save className="mr-1.5 size-4" />
              Save changes
            </Button>
          </CardFooter>
        </Card>
      </div>
    </DashboardLayout>
  )
}
