export interface GetLoginRequest {
    username: string;
    password: string;
}

export interface GetLoginResponse {
    access_token: string;
    refresh_token: string;
    token_type: string;
}

export interface GetTokenResponse {
    access_token: string;
    refresh_token: string;
    token_type: string;
}
