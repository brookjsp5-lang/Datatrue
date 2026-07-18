import type { UserRole } from './UserRole';

export interface CreateUserRequest {
  email: string;
  name: string;
  password: string;
  role: UserRole;
}
