#include <stdio.h>
#include <stdbool.h>

int main(void)
{
	int i = 2;
	float f = 3.14;
	char *s = "Hello, world!";
	bool x = true;

	printf("%s i = %d and f = %f!\n", s, i, f);

	if (x) {
		printf("x is true!\n");
	}

	int i2 = 10;
	int j2 = 5 + i2++;

	printf("%d, %d\n", i2, j2);

	i2 = 10;
	j2 = 5 + ++i2;

	printf("%d, %d\n", i2, j2);

	int a = 999;

	printf("%zu\n", sizeof a);
	printf("%zu\n", sizeof(2 + 7));
	printf("%zu\n", sizeof 3.14);
	printf("%zu\n", sizeof(int));
	printf("%zu\n", sizeof(char));
}

