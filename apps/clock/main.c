#include <ncurses.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

int main() {
	initscr();   // start ncurses mode (alternate screen)
	cbreak();    // react to ctrl+c
	noecho();    // don't print what user types
	curs_set(0); // hide the text cursor

	while (1) {
		// get the current time and apply format
		time_t t;
		struct tm *ptr;
		t = time(NULL);
		ptr = localtime(&t);
		char myTime[50];
		strftime(myTime, sizeof(myTime), "%a %b %d, %I:%M:%S %p", ptr);

		// calculate position (center in x and y axis)
		int max_y, max_x;
		getmaxyx(stdscr, max_y, max_x);
		int center_y = max_y / 2;

		// horizontal center needs an extra calculation to
		// position the text exactly in the middle considering the
		// length of the text
		int text_length = (int)strlen(myTime);
		int center_x = (max_x - text_length) / 2;

		// render the content
		clear();
		mvprintw(center_y, center_x, "%s", myTime);
		refresh();
		usleep(100000); // 100ms
	}
	endwin();
	return 0;
}
