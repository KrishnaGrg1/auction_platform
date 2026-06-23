import { login, register } from '#/lib/services/auth.services'
import { setAuth } from '#/lib/token'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { toast } from 'sonner'

export function useLogin() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setAuth(data.token)
      navigate({ to: '/dashboard' })
      toast.success(data.message)
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })
}

export function useRegister() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: register,
    onSuccess: (data) => {
      navigate({ to: '/login' })
      toast.success(data.message)
    },
    onError: (err: Error) => {
      toast.error(err.message)
    },
  })
}
