export type User = {
  id: number
  email: string
  name: string
  phone: string
  photo_url: string
  bonus_points: number
  email_verified: boolean
  phone_verified: boolean
  phone_verified_at?: string | null
  is_blocked: boolean
  is_admin: boolean
  vk_id?: string
  auth_provider?: string
  telegram_chat_id?: string
  telegram_username?: string
  telegram_linked_at?: string | null
  email_verification_sent_at?: string | null
  email_verified_at?: string | null
  created_at: string
  updated_at: string
}

export type Procedure = {
  id: number
  title: string
  description: string
  duration_mins: number
  price: number
  bonus_earned: number
  category: string
  image_url: string
  video_url: string
  services: string
  duration_str: string
  popular: boolean
  is_active: boolean
}

export type ProcedureReview = {
  id: number
  user_id: number
  user_name: string
  procedure_id: number
  procedure_title?: string
  rating: number
  text: string
  created_at: string
  updated_at: string
}

export type BeforeAfterResult = {
  id: number
  procedure_id?: number | null
  procedure: string
  title: string
  description: string
  before_url: string
  after_url: string
  is_featured: boolean
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export type Booking = {
  id: number
  user_id: number
  procedure_id: number
  datetime: string
  requested_datetimes: string[]
  comment: string
  status: string
  bonus_used: number
  notify_sms: boolean
  notify_telegram: boolean
  created_at: string
}

export type BonusLog = {
  id: number
  user_id: number
  type: string
  amount: number
  comment: string
  created_at: string
}

export type ChatLog = {
  id: number
  user_id: number | null
  user_name: string
  user_email: string
  user_phone: string
  raw_input: string
  raw_output: string
  ai_model: string
  intent: string
  created_at: string
}
