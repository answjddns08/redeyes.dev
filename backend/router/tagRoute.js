import { Router } from "express";
import { getTags } from "../controller/tagController.js";

const router = Router();

router.get("/", getTags);

export default router;
