import { DeleteResult, UpdateResult } from 'typeorm';

import { AppDataSource } from '@db/data-source';
import { User } from '@db/entity/User';

const userRepo = AppDataSource.getRepository(User);

const create = async (resource: User): Promise<User> => userRepo.save(resource);

const deleteById = async (id: string): Promise<DeleteResult> =>
  userRepo.delete(id);

const list = async (limit: number, page: number): Promise<User[]> =>
  userRepo.find({ take: limit, skip: page * limit });

const patchById = async (id: string, resource: User): Promise<UpdateResult> =>
  userRepo.update(id, resource);

const putById = async (id: string, resource: User): Promise<UpdateResult> =>
  userRepo.update(id, resource);

const readById = async (id: string): Promise<User | null> =>
  userRepo.findOneBy({ id });

const getUserByEmail = async (email: string): Promise<User | null> =>
  userRepo.findOneBy({ email });

export {
  getUserByEmail,
  readById,
  patchById,
  putById,
  list,
  deleteById,
  create,
};
