#include <sys/types.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <microhttpd.h>
#include <iostream>
#include <string>
#include <crowdsight.h>
#include <opencv2/opencv.hpp>
#include "json.hpp"

using json = nlohmann::json;

#define PORT            9999
#define POSTBUFFERSIZE  1024

enum ConnectionType
{
	GET = 0,
	POST = 1
};

CrowdSight * mCrowdSight = new CrowdSight("/usr/local/crowdsight/data/");

struct connection_info_struct
{
	enum ConnectionType connectiontype;
	struct MHD_PostProcessor *postprocessor;
	char image[2 * 1024 * 1024];
	int current;
};

const char *errorpage = "{\"error\":\"bad request\"}";

const char* const postprocerror = "{\"error\":\"error processing POST data\"}";

char * getImageData(char * image, int size)
{
	//printf("size in image data %d\n", size);
	std::vector<char> data(image, image + size);
	//printf("%ld\n", data.size());

	cv::Mat mFrame = cv::imdecode(data, 1);

	json j;

	if(mCrowdSight->process(mFrame)) 
	{
		std::vector<Person> people;
		if (!mCrowdSight->getCurrentPeople(people))
		{
			std::cerr 
				<< mCrowdSight->getErrorDescription() 
				<< std::endl;
			j["error"] = mCrowdSight->getErrorDescription();
		}

		for (int i=0; i < people.size(); i++)
		{
			j[i]["id"] = people[i].getID();
			j[i]["gender"] = people[i].getGender() < 0 ? "male" : "female";
			j[i]["age"] = people[i].getAge();
			j[i]["mood"] = people[i].getMood();
			//j[i]["pitch"] = people[i].getHeadPitch();
			//j[i]["roll"] = people[i].getHeadRoll();
			//j[i]["yaw"] = people[i].getHeadYaw();
			j[i]["emotions"] = people[i].getEmotions();
			j[i]["headpose"] = people[i].getHeadPose();
			j[i]["actionunits"] = people[i].getActionUnits();
			j[i]["headgaze"][0] = people[i].getHeadGaze().x;
			j[i]["headgaze"][1] = people[i].getHeadGaze().y;
			/*vector<float> emotions = people[i].getEmotions();
			for (int j=0; j < 6, j++)
			{
				j[i]["emotions"][j] = emotions[j];
			}*/
			
		}
		if (people.size() == 0)
		{
			j["error"] = "No faces found";
		}
	}
	else
	{
		std::cerr 
			<< mCrowdSight->getErrorDescription() 
			<< std::endl;
		j["error"] = mCrowdSight->getErrorDescription();
	}

	std::string s = j.dump();
	char * cstr = new char[s.length() + 1];
	strcpy(cstr, s.c_str());
	return cstr;
}


static int send_page (struct MHD_Connection *connection,
		const char *page,
		int status_code)
{
	int ret;
	struct MHD_Response *response;

	response =
		MHD_create_response_from_buffer (strlen (page),
				(void *) page,
				MHD_RESPMEM_MUST_COPY);
	if (!response)
		return MHD_NO;

	MHD_add_response_header (response,
			MHD_HTTP_HEADER_CONTENT_TYPE,
			"application/json");

	ret = MHD_queue_response (connection,
			status_code,
			response);
	MHD_destroy_response (response);

	return ret;
}

static int iterate_post (void *coninfo_cls,
		enum MHD_ValueKind kind,
		const char *key,
		const char *filename,
		const char *content_type,
		const char *transfer_encoding,
		const char *data,
		uint64_t off,
		size_t size)
{
	struct connection_info_struct *con_info = (connection_info_struct *) coninfo_cls;

	if (0 != strcmp (key, "file"))
		return MHD_NO;

	if (size > 0)
	{
		memcpy(con_info->image + con_info->current, data, size);
		con_info->current += size;
	}

	return MHD_YES;
}

static void request_completed (void *cls,
		struct MHD_Connection *connection,
		void **con_cls,
		enum MHD_RequestTerminationCode toe)
{
	struct connection_info_struct *con_info = (connection_info_struct *) *con_cls;

	if (NULL == con_info)
		return;

	if (con_info->connectiontype == POST)
	{
		if (NULL != con_info->postprocessor)
		{
			MHD_destroy_post_processor (con_info->postprocessor);
		}
	}

	free (con_info);
	*con_cls = NULL;
}

static int answer_to_connection (void *cls,
		struct MHD_Connection *connection,
		const char *url,
		const char *method,
		const char *version,
		const char *upload_data,
		size_t *upload_data_size,
		void **con_cls)
{
	if (NULL == *con_cls)
	{
		struct connection_info_struct *con_info;

		con_info = (connection_info_struct *) malloc (sizeof (struct connection_info_struct));
		if (NULL == con_info)
			return MHD_NO;

		con_info->current = 0;
		if (0 == strcasecmp (method, MHD_HTTP_METHOD_POST))
		{
			con_info->postprocessor =
				MHD_create_post_processor (connection,
						POSTBUFFERSIZE,
						&iterate_post,
						(void *) con_info);

			if (NULL == con_info->postprocessor)
			{
				free (con_info);
				return MHD_NO;
			}

			con_info->connectiontype = POST;
		}
		else
		{
			con_info->connectiontype = GET;
		}
		*con_cls = (void *) con_info;
		return MHD_YES;
	}

	if (0 == strcasecmp (method, MHD_HTTP_METHOD_POST))
	{
		struct connection_info_struct *con_info = (connection_info_struct *) *con_cls;

		if (0 != *upload_data_size)
		{
			if (MHD_post_process (con_info->postprocessor,
						upload_data,
						*upload_data_size) != MHD_YES)
			{
				return send_page (connection,
						postprocerror,
						MHD_HTTP_BAD_REQUEST);
			}
			*upload_data_size = 0;

			return MHD_YES;
		}
		else
		{
			char * answerstring = getImageData(con_info->image, con_info->current);
			return send_page (connection,
					answerstring,
					MHD_HTTP_OK);
		}
	}

	return send_page (connection,
			errorpage,
			MHD_HTTP_BAD_REQUEST);
}

int main ()
{
	std::string authKey = "8776d8686df14f2b84b05c7f1a951ff4";
	
	if(!mCrowdSight->authenticate(authKey))
	{
		std::cerr 
			<< "Authentication Failed : " 
			<< mCrowdSight->getErrorDescription() 
			<< std::endl;
		exit(1);
	}

	struct MHD_Daemon *daemon;

	daemon = MHD_start_daemon (MHD_USE_THREAD_PER_CONNECTION,
			PORT, NULL, NULL,
			&answer_to_connection, NULL,
			MHD_OPTION_NOTIFY_COMPLETED, &request_completed, NULL,
			MHD_OPTION_END);
	if (NULL == daemon)
		return 1;
	printf("Listing on port %d\n", PORT);
	while(1)
		(void) getchar ();
	MHD_stop_daemon (daemon);
	return 0;
}
