import { DeleteResult, UpdateResult } from 'typeorm';

export interface CRUD<T> {
  list: (limit: number, page: number) => Promise<T[]>;
  create: (resource: T) => Promise<T>;
  putById: (id: string, resource: T) => Promise<UpdateResult>;
  readById: (id: string) => Promise<T | null>;
  deleteById: (id: string) => Promise<DeleteResult>;
  patchById: (id: string, resource: T) => Promise<UpdateResult>;
}
