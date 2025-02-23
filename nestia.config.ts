import { INestiaConfig } from "@nestia/sdk";
import { NestFactory } from "@nestjs/core";
 
import { AppModule } from "./src/app.module";
 
const NESTIA_CONFIG: INestiaConfig = {
    input: async () => {
        const app = await NestFactory.create(AppModule);
        // app.setGlobalPrefix("api");
        // app.enableVersioning({
        //     type: VersioningType.URI,
        //     prefix: "v", 
        // })
        return app;
    },
    swagger: {
        output: "dist/swagger.json",
        beautify: true,
        security: {
            bearer: {
                type: "apiKey",
                name: "Authorization",
                in: "header",
            },
        },
        servers: [
            {
                url: "http://localhost:3000",
                description: "Local Server"
            }
        ],
    }
};
export default NESTIA_CONFIG;
