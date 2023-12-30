import 'reflect-metadata';
import { DataSource } from 'typeorm';

export const AppDataSource = new DataSource({
  type: 'sqlite',
  database: './gst.db',
  logging: false,
  entities: ['./src/db/entity/*'],
  migrations: ['./src/db/migration/*'],
  subscribers: [],
});
