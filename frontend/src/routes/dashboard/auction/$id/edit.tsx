import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/dashboard/auction/$id/edit')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/dashboard/auction/$id/edit"!</div>
}
