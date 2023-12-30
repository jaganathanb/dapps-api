export type Jwt = {
  refreshKey: string;
  userId: string;
  permissionLevel: string;
};

// declare module 'express' {
//   interface Response {
//     locals: any;
//   }
// }

export interface TypedRequest<T, U, P> extends Express.Request {
  body: T;
  query: U;
  params: P;
}

export interface TypedResponse extends Express.Response {
  locals: Jwt | string | { errorMessage: string };
}
