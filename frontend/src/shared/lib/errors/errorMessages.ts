// Error message mappings for better user experience
export const ERROR_MESSAGES: Record<string, string> = {
  // Authentication errors
  'USERNAME_ALREADY_EXISTS': 'This username is already taken. Please choose a different one.',
  'EMAIL_ALREADY_EXISTS': 'An account with this email already exists. Please use a different email or try logging in.',
  'INVALID_CREDENTIALS': 'Invalid username or password. Please check your credentials and try again.',
  'INVALID_REQUEST_DATA': 'Please fill in all required fields correctly.',
  
  // Generic errors
  'NETWORK_ERROR': 'Network error. Please check your internet connection and try again.',
  'SERVER_ERROR': 'Server error. Please try again later.',
  'VALIDATION_ERROR': 'Please check your input and try again.',
  
  // Password validation
  'PASSWORD_TOO_SHORT': 'Password must be at least 6 characters long.',
  'PASSWORD_TOO_WEAK': 'Password must contain at least one uppercase letter, one lowercase letter, and one number.',
  
  // Email validation
  'INVALID_EMAIL': 'Please enter a valid email address.',
  
  // Username validation
  'USERNAME_TOO_SHORT': 'Username must be at least 3 characters long.',
  'USERNAME_TOO_LONG': 'Username must be no more than 50 characters long.',
  'USERNAME_INVALID_CHARS': 'Username can only contain letters, numbers, and underscores.',
};

export function getErrorMessage(errorCode: string): string {
  return ERROR_MESSAGES[errorCode] || errorCode || 'An unexpected error occurred. Please try again.';
}

export function isKnownError(errorCode: string): boolean {
  return errorCode in ERROR_MESSAGES;
}
