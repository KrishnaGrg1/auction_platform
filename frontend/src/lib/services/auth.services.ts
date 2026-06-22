import { createServerFn } from '@tanstack/react-start'
import { authClient } from '../api/auth.api'
import { LoginUserSchema, RegisterUserSchema } from '../schema/auth.schema'
import { ConnectError } from '@connectrpc/connect'


export const register = createServerFn({ method: 'POST' })
  .validator((data) => RegisterUserSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      const res = await authClient.register({
        email: data.email,
        firstName: data.firstName,
        lastName: data.lastName,
        password: data.password,
      })
      return res
    } catch (error: unknown) {
      if (error instanceof ConnectError) {
        const msg =
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message

        throw new Error(msg)
      }

      throw new Error('Failed to login')
    }
  })

export const login = createServerFn({ method: 'POST' })
  .validator((data) => LoginUserSchema.parse(data))
  .handler(async ({ data }) => {
    try {
      const res = await authClient.login({
        email: data.email,
        password: data.password,
      })
      console.log("data",res)
      return res
    } catch (error: unknown) {
      console.log("error ho ",error)
      if (error instanceof ConnectError) {
        const msg =
          typeof error.rawMessage === 'string'
            ? error.rawMessage
            : error.message
        throw new Error(msg)
      }

      throw new Error('Failed to login')
    }
  })


