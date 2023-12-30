import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

export type UserRoleType = 'admin' | 'user' | 'guest';

@Entity()
export class User {
  @PrimaryGeneratedColumn('uuid')
  public id!: string;

  @Column()
  public firstName!: string;

  @Column()
  public lastName!: string;

  @Column()
  public age!: number;

  @Column()
  public email!: string;

  @Column()
  public password!: string;

  @Column({
    type: 'simple-enum',
    enum: ['admin', 'editor', 'ghost'],
    default: 'user',
  })
  public role!: UserRoleType;
}
