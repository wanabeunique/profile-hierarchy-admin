export interface Profile {
  id: number
  name: string
  email: string
  type: 'plus' | 'standard' | 'min'
  commission: number
  paymentDeadline: string | null
  parentId: number | null
  parent?: Profile | null
  children?: Profile[]
  orders?: Order[]
  createdAt: string
  orderCount: number
  totalAmount: number
  totalPaid: number
  totalRemaining: number
  subAccountCount: number
}

export interface Order {
  id: number
  profileId: number
  number: string
  amount: number
  paid: number
  remaining: number
  status: 'Не оплачено' | 'Частично оплачено' | 'Оплачено'
  createdAt: string
}

export interface LoginResponse {
  token: string
}

// Matches backend JWTClaims: json tags are profile_id, profile_type
export interface JwtPayload {
  role: 'admin' | 'profile'
  profile_id?: number
  profile_type?: 'plus' | 'standard' | 'min'
  exp: number
}
