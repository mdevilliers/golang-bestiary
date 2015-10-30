#include "calc-shared.h"
#include <stdio.h>

int main() {

	printf("%s\n", "This is a C application.");
	GoInt a = 10;
	GoInt b = 20;

	GoInt c = Sum(a, b);

	printf("Result from Go : %d\n", (int)c);

	return 0;

}