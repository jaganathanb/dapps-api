import { MigrationInterface, QueryRunner } from "typeorm";

export class AddedPasswordColumn1703959332208 implements MigrationInterface {
    name = 'AddedPasswordColumn1703959332208'

    public async up(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`CREATE TABLE "temporary_user" ("id" varchar PRIMARY KEY NOT NULL, "firstName" varchar NOT NULL, "lastName" varchar NOT NULL, "age" integer NOT NULL, "email" varchar NOT NULL, "role" varchar CHECK( "role" IN ('admin','editor','ghost') ) NOT NULL DEFAULT ('user'), "password" varchar NOT NULL)`);
        await queryRunner.query(`INSERT INTO "temporary_user"("id", "firstName", "lastName", "age", "email", "role") SELECT "id", "firstName", "lastName", "age", "email", "role" FROM "user"`);
        await queryRunner.query(`DROP TABLE "user"`);
        await queryRunner.query(`ALTER TABLE "temporary_user" RENAME TO "user"`);
    }

    public async down(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`ALTER TABLE "user" RENAME TO "temporary_user"`);
        await queryRunner.query(`CREATE TABLE "user" ("id" varchar PRIMARY KEY NOT NULL, "firstName" varchar NOT NULL, "lastName" varchar NOT NULL, "age" integer NOT NULL, "email" varchar NOT NULL, "role" varchar CHECK( "role" IN ('admin','editor','ghost') ) NOT NULL DEFAULT ('user'))`);
        await queryRunner.query(`INSERT INTO "user"("id", "firstName", "lastName", "age", "email", "role") SELECT "id", "firstName", "lastName", "age", "email", "role" FROM "temporary_user"`);
        await queryRunner.query(`DROP TABLE "temporary_user"`);
    }

}
