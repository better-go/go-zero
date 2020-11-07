
########################################################

git-add:
	git add .

# gitmoji comments:
git-commit:
	git commit -m ":tada: :bulb: :rocket: TODO X: update comments"

git-push:
	git push origin annotated-v1.0.25

# quick push:
quick: git-commit git-push


