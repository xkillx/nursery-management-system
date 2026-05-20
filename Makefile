.PHONY: seed

seed:
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" && exit 1)
	@test -n "$$SEED_EMAIL" || (echo "SEED_EMAIL is required" && exit 1)
	@test -n "$$SEED_PASSWORD" || (echo "SEED_PASSWORD is required" && exit 1)
	@cd api && go run ./cmd/seed \
	  --email "$$SEED_EMAIL" \
	  --password "$$SEED_PASSWORD" \
	  --tenant "${SEED_TENANT:-Pilot Nursery}" \
	  --branch "${SEED_BRANCH:-Main}"
