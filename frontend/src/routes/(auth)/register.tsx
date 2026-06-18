import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from '@tanstack/react-form'
import { Eye, EyeOff, Gavel } from 'lucide-react'
import { useRegister } from '#/hooks/use-auth'

export const Route = createFileRoute('/(auth)/register')({
  component: RegisterPage,
})

function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false)
  const { mutate: register, isPending, error: registerError } = useRegister()

  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
    },
    onSubmit: async ({ value }) => {
      register({ data: value })
    },
  })

  return (
    <div
      className="flex min-h-screen"
      style={{
        background: 'var(--surface)',
        fontFamily: 'var(--font-heading)',
      }}
    >
      {/* ── Left panel — brand ── */}
      <div
        className="hidden lg:flex lg:w-[42%] flex-col justify-between p-10 xl:p-14"
        style={{ background: 'var(--ink)' }}
      >
        {/* Logo */}
        <Link to="/" className="flex items-center gap-3 no-underline">
          <span
            className="flex size-9 items-center justify-center rounded-lg"
            style={{ background: 'var(--amber)' }}
          >
            <Gavel className="size-4" style={{ color: 'var(--ink)' }} />
          </span>
          <span
            style={{
              fontFamily: 'var(--font-heading)',
              fontWeight: 600,
              fontSize: '0.9375rem',
              color: 'var(--amber-light)',
            }}
          >
            Auction Platform
          </span>
        </Link>

        {/* Middle */}
        <div className="flex flex-col gap-8">
          <h2
            style={{
              fontFamily: 'var(--font-heading)',
              fontSize: 'clamp(1.6rem, 2.5vw, 2.25rem)',
              fontWeight: 700,
              lineHeight: 1.15,
              color: 'var(--amber-light)',
              letterSpacing: '-0.02em',
            }}
          >
            Start selling
            <br />
            and bidding today.
          </h2>

          <ul className="flex flex-col gap-3">
            {[
              'Create auction listings',
              'Track bids in real time',
              'Manage buyer activity',
            ].map((item) => (
              <li key={item} className="flex items-center gap-3">
                <span
                  className="flex size-5 shrink-0 items-center justify-center rounded-full"
                  style={{
                    background: 'var(--amber-bg)',
                    border: '0.5px solid var(--amber-border)',
                  }}
                >
                  <span
                    className="size-1.5 rounded-full"
                    style={{ background: 'var(--amber)' }}
                  />
                </span>
                <span
                  style={{
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.8125rem',
                    color: 'rgba(250,248,244,0.60)',
                  }}
                >
                  {item}
                </span>
              </li>
            ))}
          </ul>
        </div>

        {/* Footer */}
        <p
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.6875rem',
            color: 'rgba(250,248,244,0.22)',
          }}
        >
          One account for bidding, selling, and settlement
        </p>
      </div>

      {/* ── Right panel — form ── */}
      <div className="flex flex-1 flex-col items-center justify-center px-6 py-12">
        {/* Mobile logo */}
        <Link
          to="/"
          className="flex items-center gap-2.5 no-underline mb-10 lg:hidden"
        >
          <span
            className="flex size-8 items-center justify-center rounded-lg"
            style={{ background: 'var(--ink)' }}
          >
            <Gavel
              className="size-3.5"
              style={{ color: 'var(--amber-light)' }}
            />
          </span>
          <span
            style={{
              fontFamily: 'var(--font-heading)',
              fontWeight: 600,
              fontSize: '0.875rem',
              color: 'var(--ink)',
            }}
          >
            Auction Platform
          </span>
        </Link>

        <div className="rise-in w-full max-w-[420px]">
          {/* Heading */}
          <div className="mb-7">
            <h1
              style={{
                fontFamily: 'var(--font-heading)',
                fontSize: '1.5rem',
                fontWeight: 700,
                color: 'var(--ink)',
                letterSpacing: '-0.015em',
                marginBottom: '0.375rem',
              }}
            >
              Create account
            </h1>
            <p
              style={{
                fontSize: '0.875rem',
                color: 'var(--ink-soft)',
                lineHeight: 1.6,
              }}
            >
              Join the auction platform to list items and place bids.
            </p>
          </div>

          {/* Error */}
          {registerError && (
            <div
              className="mb-5 rounded-lg px-4 py-3 text-sm"
              style={{
                background: '#fef2f2',
                border: '0.5px solid #fecaca',
                color: '#b91c1c',
                fontFamily: 'var(--font-heading)',
              }}
            >
              {registerError instanceof Error
                ? registerError.message
                : 'Unable to create account.'}
            </div>
          )}

          {/* Form */}
          <form
            onSubmit={(e) => {
              e.preventDefault()
              e.stopPropagation()
              form.handleSubmit()
            }}
            className="flex flex-col gap-4"
          >
            {/* First + Last name row */}
            <div className="grid grid-cols-2 gap-3">
              <form.Field
                name="firstName"
                validators={{
                  onChange: ({ value }) =>
                    !value.trim() ? 'Required' : undefined,
                }}
              >
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <label
                      htmlFor={field.name}
                      style={{
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        color: 'var(--ink)',
                        fontFamily: 'var(--font-heading)',
                        letterSpacing: '0.01em',
                      }}
                    >
                      First name
                    </label>
                    <FieldInput
                      id={field.name}
                      name={field.name}
                      type="text"
                      value={field.state.value}
                      placeholder="Aarav"
                      hasError={field.state.meta.errors.length > 0}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                    />
                    {field.state.meta.errors.length > 0 && (
                      <p
                        style={{
                          fontSize: '0.75rem',
                          color: '#dc2626',
                          fontFamily: 'var(--font-heading)',
                        }}
                      >
                        {field.state.meta.errors[0]}
                      </p>
                    )}
                  </div>
                )}
              </form.Field>

              <form.Field
                name="lastName"
                validators={{
                  onChange: ({ value }) =>
                    !value.trim() ? 'Required' : undefined,
                }}
              >
                {(field) => (
                  <div className="flex flex-col gap-1.5">
                    <label
                      htmlFor={field.name}
                      style={{
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        color: 'var(--ink)',
                        fontFamily: 'var(--font-heading)',
                        letterSpacing: '0.01em',
                      }}
                    >
                      Last name
                    </label>
                    <FieldInput
                      id={field.name}
                      name={field.name}
                      type="text"
                      value={field.state.value}
                      placeholder="Sharma"
                      hasError={field.state.meta.errors.length > 0}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                    />
                    {field.state.meta.errors.length > 0 && (
                      <p
                        style={{
                          fontSize: '0.75rem',
                          color: '#dc2626',
                          fontFamily: 'var(--font-heading)',
                        }}
                      >
                        {field.state.meta.errors[0]}
                      </p>
                    )}
                  </div>
                )}
              </form.Field>
            </div>

            {/* Email */}
            <form.Field
              name="email"
              validators={{
                onChange: ({ value }) =>
                  !value
                    ? 'Email is required'
                    : !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)
                      ? 'Invalid email format'
                      : undefined,
              }}
            >
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor={field.name}
                    style={{
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: 'var(--ink)',
                      fontFamily: 'var(--font-heading)',
                      letterSpacing: '0.01em',
                    }}
                  >
                    Email
                  </label>
                  <FieldInput
                    id={field.name}
                    name={field.name}
                    type="email"
                    value={field.state.value}
                    placeholder="you@example.com"
                    hasError={field.state.meta.errors.length > 0}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                  />
                  {field.state.meta.errors.length > 0 && (
                    <p
                      style={{
                        fontSize: '0.75rem',
                        color: '#dc2626',
                        fontFamily: 'var(--font-heading)',
                      }}
                    >
                      {field.state.meta.errors[0]}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            {/* Password */}
            <form.Field
              name="password"
              validators={{
                onChange: ({ value }) =>
                  !value
                    ? 'Password is required'
                    : value.length < 8
                      ? 'At least 8 characters'
                      : undefined,
              }}
            >
              {(field) => (
                <div className="flex flex-col gap-1.5">
                  <label
                    htmlFor={field.name}
                    style={{
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      color: 'var(--ink)',
                      fontFamily: 'var(--font-heading)',
                      letterSpacing: '0.01em',
                    }}
                  >
                    Password
                  </label>
                  <div className="relative">
                    <FieldInput
                      id={field.name}
                      name={field.name}
                      type={showPassword ? 'text' : 'password'}
                      value={field.state.value}
                      placeholder="••••••••"
                      hasError={field.state.meta.errors.length > 0}
                      paddingRight
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
                      style={{ color: 'var(--ink-muted)' }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.color = 'var(--ink)')
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.color = 'var(--ink-muted)')
                      }
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                  {field.state.meta.errors.length > 0 && (
                    <p
                      style={{
                        fontSize: '0.75rem',
                        color: '#dc2626',
                        fontFamily: 'var(--font-heading)',
                      }}
                    >
                      {field.state.meta.errors[0]}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            {/* Submit */}
            <button
              type="submit"
              disabled={isPending}
              className="mt-1 w-full rounded-lg py-2.5 text-sm font-semibold transition-opacity hover:opacity-90 disabled:opacity-60"
              style={{
                background: 'var(--ink)',
                color: 'var(--amber-light)',
                fontFamily: 'var(--font-heading)',
                border: 'none',
                cursor: isPending ? 'not-allowed' : 'pointer',
              }}
            >
              {isPending ? 'Creating account…' : 'Create account →'}
            </button>
          </form>

          {/* Footer */}
          <p
            className="mt-6 text-center text-xs"
            style={{
              color: 'var(--ink-muted)',
              fontFamily: 'var(--font-heading)',
            }}
          >
            Already registered?{' '}
            <Link
              to="/login"
              style={{
                color: 'var(--amber)',
                fontWeight: 600,
                textDecoration: 'none',
              }}
            >
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}

/* ── Shared input component — avoids repeating inline styles ── */
function FieldInput({
  hasError,
  paddingRight,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & {
  hasError: boolean
  paddingRight?: boolean
}) {
  return (
    <input
      {...props}
      style={{
        fontFamily: 'var(--font-heading)',
        fontSize: '0.9rem',
        color: 'var(--ink)',
        background: 'var(--card-bg)',
        border: hasError
          ? '1px solid #fca5a5'
          : '0.5px solid var(--line-strong)',
        borderRadius: '0.5rem',
        padding: paddingRight
          ? '0.625rem 2.5rem 0.625rem 0.875rem'
          : '0.625rem 0.875rem',
        width: '100%',
        outline: 'none',
        transition: 'border-color 0.15s',
      }}
      onFocus={(e) => {
        if (!hasError) {
          e.target.style.borderColor = 'var(--amber)'
          e.target.style.borderWidth = '1px'
        }
        props.onFocus?.(e)
      }}
      onBlur={(e) => {
        if (!hasError) {
          e.target.style.borderColor = 'var(--line-strong)'
          e.target.style.borderWidth = '0.5px'
        }
        props.onBlur?.(e)
      }}
    />
  )
}
