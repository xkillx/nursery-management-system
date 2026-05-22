export interface ApiErrorBody {
  code: string;
  message: string;
  details?: unknown;
  request_id?: string;
}

export interface MappedApiError {
  code: string;
  message: string;
  requestId: string | null;
  fieldErrors: Record<string, string>;
}
