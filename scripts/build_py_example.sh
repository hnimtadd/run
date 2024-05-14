cd examples/python &&
	rm -rf bundle dist &&
	poetry install --no-root &&
	poetry build -o build/dist &&
	poetry run pip install --upgrade -t build/bundle build/dist/*.whl &&
	cd build/bundle &&
	py2wasm -o ../../index.wasm run/lambda_handler/handler.py &&
	cd ../../../../../../
