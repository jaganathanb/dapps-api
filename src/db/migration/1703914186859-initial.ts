import { MigrationInterface, QueryRunner } from 'typeorm';

export class Initial1703914186859 implements MigrationInterface {
  public name = 'Initial1703914186859';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "user" ("id" varchar PRIMARY KEY NOT NULL, "firstName" varchar NOT NULL, "lastName" varchar NOT NULL, 
      "age" integer NOT NULL, "email" varchar NOT NULL)`,
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "user"`);
  }
}
