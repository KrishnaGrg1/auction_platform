import { z } from 'zod'

export const RegisterUserSchema = z.object({
  email: z.string().email(),
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  password: z.string().min(8),
})

export const LoginUserSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
})