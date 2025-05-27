export interface PasswordValidation {
  isValid: boolean;
  errors: string[];
}

export const validatePassword = (password: string): PasswordValidation => {
  const errors: string[] = [];

  // Минимальная длина 8 символов
  if (password.length < 8) {
    errors.push('Пароль должен содержать минимум 8 символов');
  }

  // Проверка на наличие заглавной буквы
  if (!/[A-Z]/.test(password)) {
    errors.push('Пароль должен содержать минимум одну заглавную букву');
  }

  // Проверка на наличие строчной буквы
  if (!/[a-z]/.test(password)) {
    errors.push('Пароль должен содержать минимум одну строчную букву');
  }

  // Проверка на наличие цифры
  if (!/\d/.test(password)) {
    errors.push('Пароль должен содержать минимум одну цифру');
  }

  // Проверка на наличие специального символа
  if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
    errors.push('Пароль должен содержать минимум один специальный символ');
  }

  return {
    isValid: errors.length === 0,
    errors
  };
};

export const getPasswordStrength = (password: string): 'weak' | 'medium' | 'strong' => {
  const validation = validatePassword(password);
  
  if (password.length === 0) return 'weak';
  
  let score = 0;
  
  if (password.length >= 8) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[a-z]/.test(password)) score++;
  if (/\d/.test(password)) score++;
  if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) score++;
  
  if (score < 3) return 'weak';
  if (score < 5) return 'medium';
  return 'strong';
};
