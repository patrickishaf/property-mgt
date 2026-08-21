import { useState } from 'react';
import {
  Building,
  Sun,
  Moon,
  Lock,
  Phone,
  KeyRound,
  ArrowLeft,
  Loader2,
  AlertCircle,
  CheckCircle2,
  Mail,
} from 'lucide-react';
import { useApp } from '../context';

type Method = 'password' | 'otp';
type OtpStep = 'request' | 'verify';

export function Login() {
  const { loginWithPassword, sendOtp, verifyOtp, authLoading, theme, toggleTheme } = useApp();
  const [method, setMethod] = useState<Method>('password');
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [otpStep, setOtpStep] = useState<OtpStep>('request');
  const [otpCode, setOtpCode] = useState('');
  const [sentCode, setSentCode] = useState('');
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');

  const reset = () => {
    setError('');
    setInfo('');
    setOtpStep('request');
    setOtpCode('');
    setSentCode('');
  };

  const switchMethod = (m: Method) => {
    setMethod(m);
    reset();
    setPassword('');
  };

  const handlePasswordLogin = async () => {
    if (!identifier.trim()) {
      setError('Enter your email or phone number.');
      return;
    }
    if (!password) {
      setError('Enter your password.');
      return;
    }
    setError('');
    setInfo('');
    try {
      await loginWithPassword(identifier, password);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login failed. Please try again.');
    }
  };

  const handleSendOtp = async () => {
    if (!identifier.trim()) {
      setError('Enter your email or phone number.');
      return;
    }
    setError('');
    setInfo('');
    try {
      const code = await sendOtp(identifier);
      setSentCode(code);
      setOtpStep('verify');
      setInfo(`A one-time code has been sent. For this demo, your code is: ${code}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to send code. Please try again.');
    }
  };

  const handleVerifyOtp = async () => {
    if (!otpCode.trim()) {
      setError('Enter the code you received.');
      return;
    }
    setError('');
    setInfo('');
    try {
      await verifyOtp(identifier, otpCode);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Verification failed. Please try again.');
    }
  };

  const resendCode = async () => {
    setOtpCode('');
    setError('');
    await handleSendOtp();
  };

  return (
    <div className="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
      <div className="absolute top-4 right-4">
        <button onClick={toggleTheme} className="btn-ghost p-2" aria-label="Toggle theme">
          {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
        </button>
      </div>

      <div className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          {/* Brand */}
          <div className="flex flex-col items-center mb-8">
            <div className="w-12 h-12 bg-brand-600 flex items-center justify-center mb-4">
              <Building size={24} className="text-white" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight">RentLedger</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Property & Rent Management Dashboard
            </p>
          </div>

          {/* Login card */}
          <div className="surface p-6 sm:p-8">
            {/* Method toggle */}
            <div className="flex border border-gray-200 dark:border-gray-800 mb-6">
              <button
                onClick={() => switchMethod('password')}
                className={`flex-1 flex items-center justify-center gap-2 py-2.5 text-sm font-medium transition-colors ${
                  method === 'password'
                    ? 'bg-brand-600 text-white'
                    : 'text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
              >
                <Lock size={15} /> Password
              </button>
              <button
                onClick={() => switchMethod('otp')}
                className={`flex-1 flex items-center justify-center gap-2 py-2.5 text-sm font-medium transition-colors ${
                  method === 'otp'
                    ? 'bg-brand-600 text-white'
                    : 'text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
              >
                <KeyRound size={15} /> One-time code
              </button>
            </div>

            {/* Error */}
            {error && (
              <div className="flex items-start gap-2 px-3 py-2.5 text-sm text-red-700 bg-red-50 border border-red-200 dark:text-red-300 dark:bg-red-950 dark:border-red-800 mb-4">
                <AlertCircle size={16} className="shrink-0 mt-0.5" />
                <span>{error}</span>
              </div>
            )}

            {/* Info (OTP sent) */}
            {info && method === 'otp' && otpStep === 'verify' && (
              <div className="flex items-start gap-2 px-3 py-2.5 text-sm text-emerald-700 bg-emerald-50 border border-emerald-200 dark:text-emerald-300 dark:bg-emerald-950 dark:border-emerald-800 mb-4">
                <CheckCircle2 size={16} className="shrink-0 mt-0.5" />
                <span>{info}</span>
              </div>
            )}

            {/* Identifier */}
            <div className="mb-4">
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Email or Phone
              </label>
              <div className="relative">
                <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  className="input-base w-full pl-9 pr-3 py-2.5 text-sm"
                  placeholder="you@example.com or +233 24 000 0000"
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  disabled={method === 'otp' && otpStep === 'verify'}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      if (method === 'password') handlePasswordLogin();
                      else if (otpStep === 'request') handleSendOtp();
                    }
                  }}
                />
              </div>
            </div>

            {/* Password method */}
            {method === 'password' && (
              <div className="mb-6">
                <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                  Password
                </label>
                <div className="relative">
                  <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    type="password"
                    className="input-base w-full pl-9 pr-3 py-2.5 text-sm"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handlePasswordLogin()}
                  />
                </div>
              </div>
            )}

            {/* OTP verify step */}
            {method === 'otp' && otpStep === 'verify' && (
              <>
                <div className="mb-6">
                  <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                    One-time Code
                  </label>
                  <div className="relative">
                    <KeyRound size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      type="text"
                      inputMode="numeric"
                      maxLength={6}
                      className="input-base w-full pl-9 pr-3 py-2.5 text-sm font-mono tracking-widest"
                      placeholder="000000"
                      value={otpCode}
                      onChange={(e) => setOtpCode(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && handleVerifyOtp()}
                    />
                  </div>
                </div>
                <button
                  onClick={reset}
                  className="flex items-center gap-1.5 text-xs text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 mb-4"
                >
                  <ArrowLeft size={13} /> Use a different email or phone
                </button>
              </>
            )}

            {/* Action button */}
            {method === 'password' && (
              <button
                onClick={handlePasswordLogin}
                disabled={authLoading}
                className="btn btn-primary w-full"
              >
                {authLoading ? <Loader2 size={16} className="animate-spin" /> : <Lock size={16} />}
                {authLoading ? 'Signing in…' : 'Sign In'}
              </button>
            )}

            {method === 'otp' && otpStep === 'request' && (
              <button
                onClick={handleSendOtp}
                disabled={authLoading}
                className="btn btn-primary w-full"
              >
                {authLoading ? <Loader2 size={16} className="animate-spin" /> : <Phone size={16} />}
                {authLoading ? 'Sending code…' : 'Send Code'}
              </button>
            )}

            {method === 'otp' && otpStep === 'verify' && (
              <button
                onClick={handleVerifyOtp}
                disabled={authLoading}
                className="btn btn-primary w-full"
              >
                {authLoading ? <Loader2 size={16} className="animate-spin" /> : <KeyRound size={16} />}
                {authLoading ? 'Verifying…' : 'Verify & Sign In'}
              </button>
            )}

            {/* Resend link */}
            {method === 'otp' && otpStep === 'verify' && (
              <p className="text-center text-xs text-gray-400 mt-4">
                Didn't receive it?{' '}
                <button
                  onClick={resendCode}
                  className="text-brand-600 dark:text-brand-400 font-medium hover:underline"
                >
                  Resend code
                </button>
              </p>
            )}
          </div>

          {/* Demo hint */}
          <div className="mt-6 surface-muted p-4 text-xs text-gray-500 dark:text-gray-400">
            <p className="font-semibold text-gray-600 dark:text-gray-300 mb-2">Demo Accounts</p>
            <div className="space-y-1">
              <p><span className="text-gray-400">Owner:</span> daniel.okoro@rentledger.com / owner123</p>
              <p><span className="text-gray-400">Manager:</span> grace.mensah@rentledger.com / manager123</p>
              <p><span className="text-gray-400">Tenant:</span> kwame.b@mail.com / tenant123</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
